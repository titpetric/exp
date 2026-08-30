package coverage

import (
	"testing"

	"github.com/titpetric/exp/cmd/go-fsck/model"
)

func TestModulePath(t *testing.T) {
	def := func(module string) *model.Definition {
		if module == "" {
			return &model.Definition{}
		}
		return &model.Definition{Module: &model.Module{Path: module}}
	}

	for name, test := range map[string]struct {
		defs []*model.Definition
		want string
	}{
		"none":               {nil, ""},
		"no go.mod":          {[]*model.Definition{def("")}, ""},
		"one module":         {[]*model.Definition{def("example.com/a"), def("example.com/a")}, "example.com/a"},
		"module and no mod":  {[]*model.Definition{def(""), def("example.com/a")}, "example.com/a"},
		"two modules answer": {[]*model.Definition{def("example.com/a"), def("example.com/b")}, ""},
	} {
		t.Run(name, func(t *testing.T) {
			if got := modulePath(test.defs); got != test.want {
				t.Errorf("modulePath() = %q, want %q", got, test.want)
			}
		})
	}
}

func TestRelativePackage(t *testing.T) {
	for name, test := range map[string]struct {
		module     string
		importPath string
		want       string
	}{
		"root":            {"example.com/a", "example.com/a", "."},
		"below root":      {"example.com/a", "example.com/a/cmd/tool", "cmd/tool"},
		"no module":       {"", "example.com/a/cmd/tool", "example.com/a/cmd/tool"},
		"outside module":  {"example.com/a", "example.com/b/cmd", "example.com/b/cmd"},
		"shared prefix":   {"example.com/a", "example.com/ab/cmd", "example.com/ab/cmd"},
		"module is empty": {"", "", ""},
	} {
		t.Run(name, func(t *testing.T) {
			if got := relativePackage(test.module, test.importPath); got != test.want {
				t.Errorf("relativePackage(%q, %q) = %q, want %q", test.module, test.importPath, got, test.want)
			}
		})
	}
}

func TestCollapseColumn(t *testing.T) {
	rows := [][]string{
		{"ok", "a", "one"},
		{"ok", "a", "two"},
		{"ok", "b", "three"},
		{"ok", "b", "four"},
		{"ok", "a", "five"},
	}
	collapseColumn(rows, 1)

	// The fifth row repeats "a" after "b", which is a new group rather than a
	// continuation of the first, so it is written out again.
	want := []string{"a", "", "b", "", "a"}
	for i, row := range rows {
		if row[1] != want[i] {
			t.Errorf("row %d package = %q, want %q", i, row[1], want[i])
		}
	}

	// Nothing else in the row moves.
	if rows[4][2] != "five" || rows[0][0] != "ok" {
		t.Errorf("collapseColumn touched a column it was not given: %v", rows)
	}
}

func TestCollapseColumnShortRow(t *testing.T) {
	rows := [][]string{{"a"}, {"a", "b"}, {"a", "b"}}
	collapseColumn(rows, 1)
	if rows[1][1] != "b" || rows[2][1] != "" {
		t.Errorf("collapseColumn over a short row = %v", rows)
	}
}
