package diff

import (
	"reflect"
	"testing"

	"github.com/titpetric/exp/cmd/go-fsck/model"
)

// modDefs builds the definitions a model extracted from one module produces,
// which is a package carrying the go.mod it belongs to.
func modDefs(module *model.Module) []*model.Definition {
	d := def("example.com/x", "x", false, nil, nil)
	d.Module = module
	return []*model.Definition{d}
}

// mod builds a go.mod holding the given requirements.
func mod(requires ...model.Require) *model.Module {
	return &model.Module{Path: "example.com/x", GoVersion: "1.24.0", Requires: requires}
}

func TestCompareModules(t *testing.T) {
	tests := []struct {
		title           string
		old             *model.Module
		new             *model.Module
		includeIndirect bool
		want            []ModuleChange
	}{
		{
			title: "no change",
			old:   mod(model.Require{Path: "example.com/a", Version: "v1.0.0"}),
			new:   mod(model.Require{Path: "example.com/a", Version: "v1.0.0"}),
			want:  []ModuleChange{},
		},
		{
			title: "a requirement moved to another version",
			old:   mod(model.Require{Path: "example.com/a", Version: "v1.0.0"}),
			new:   mod(model.Require{Path: "example.com/a", Version: "v1.1.0"}),
			want: []ModuleChange{{
				Path: "example.com/x",
				Requires: []RequireChange{{
					Path:   "example.com/a",
					Change: requireChanged,
					Old:    &Require{Path: "example.com/a", Version: "v1.0.0"},
					New:    &Require{Path: "example.com/a", Version: "v1.1.0"},
				}},
			}},
		},
		{
			title: "a requirement was taken on",
			old:   mod(),
			new:   mod(model.Require{Path: "example.com/a", Version: "v1.0.0"}),
			want: []ModuleChange{{
				Path: "example.com/x",
				Requires: []RequireChange{{
					Path:   "example.com/a",
					Change: requireAdded,
					New:    &Require{Path: "example.com/a", Version: "v1.0.0"},
				}},
			}},
		},
		{
			title: "a requirement was dropped",
			old:   mod(model.Require{Path: "example.com/a", Version: "v1.0.0"}),
			new:   mod(),
			want: []ModuleChange{{
				Path: "example.com/x",
				Requires: []RequireChange{{
					Path:   "example.com/a",
					Change: requireRemoved,
					Old:    &Require{Path: "example.com/a", Version: "v1.0.0"},
				}},
			}},
		},
		{
			title: "an indirect requirement is left out",
			old:   mod(model.Require{Path: "example.com/a", Version: "v1.0.0", Indirect: true}),
			new:   mod(model.Require{Path: "example.com/a", Version: "v1.1.0", Indirect: true}),
			want:  []ModuleChange{},
		},
		{
			title:           "an indirect requirement is reported when asked for",
			old:             mod(model.Require{Path: "example.com/a", Version: "v1.0.0", Indirect: true}),
			new:             mod(model.Require{Path: "example.com/a", Version: "v1.1.0", Indirect: true}),
			includeIndirect: true,
			want: []ModuleChange{{
				Path: "example.com/x",
				Requires: []RequireChange{{
					Path:   "example.com/a",
					Change: requireChanged,
					Old:    &Require{Path: "example.com/a", Version: "v1.0.0", Indirect: true},
					New:    &Require{Path: "example.com/a", Version: "v1.1.0", Indirect: true},
				}},
			}},
		},
		{
			title: "a requirement that became direct is reported",
			old:   mod(model.Require{Path: "example.com/a", Version: "v1.0.0", Indirect: true}),
			new:   mod(model.Require{Path: "example.com/a", Version: "v1.0.0"}),
			want: []ModuleChange{{
				Path: "example.com/x",
				Requires: []RequireChange{{
					Path:   "example.com/a",
					Change: requireChanged,
					Old:    &Require{Path: "example.com/a", Version: "v1.0.0", Indirect: true},
					New:    &Require{Path: "example.com/a", Version: "v1.0.0"},
				}},
			}},
		},
		{
			title: "the go directive moved",
			old:   &model.Module{Path: "example.com/x", GoVersion: "1.24.0"},
			new:   &model.Module{Path: "example.com/x", GoVersion: "1.27.0"},
			want: []ModuleChange{{
				Path: "example.com/x",
				Go:   &VersionChange{Old: "1.24.0", New: "1.27.0"},
			}},
		},
		{
			title: "a toolchain was pinned",
			old:   &model.Module{Path: "example.com/x"},
			new:   &model.Module{Path: "example.com/x", Toolchain: "go1.27.0"},
			want: []ModuleChange{{
				Path:      "example.com/x",
				Toolchain: &VersionChange{New: "go1.27.0"},
			}},
		},
		{
			title: "a replace was taken on",
			old:   &model.Module{Path: "example.com/x"},
			new: &model.Module{Path: "example.com/x", Replaces: []model.Replace{
				{Path: "example.com/a", NewPath: "example.com/fork/a", NewVersion: "v1.2.0"},
			}},
			want: []ModuleChange{{
				Path: "example.com/x",
				Replaces: []ReplaceChange{{
					Path:   "example.com/a",
					Change: requireAdded,
					New:    &Replace{Path: "example.com/a", NewPath: "example.com/fork/a", NewVersion: "v1.2.0"},
				}},
			}},
		},
	}

	for _, test := range tests {
		got := Compare(modDefs(test.old), modDefs(test.new), false, test.includeIndirect)
		if !reflect.DeepEqual(got.Modules, test.want) {
			t.Errorf("%s: Compare().Modules = %#v, want %#v", test.title, got.Modules, test.want)
		}
		if got.Breaking {
			t.Errorf("%s: Compare() called a dependency change breaking", test.title)
		}
	}
}

func TestCompareModulesReportsAModuleOnlyOneRevisionHas(t *testing.T) {
	// A release that splits a module out has no older go.mod to compare it
	// against, so everything it requires reads as taken on.
	got := Compare(nil, modDefs(mod(model.Require{Path: "example.com/a", Version: "v1.0.0"})), false, false)

	want := []ModuleChange{{
		Path: "example.com/x",
		Go:   &VersionChange{New: "1.24.0"},
		Requires: []RequireChange{{
			Path:   "example.com/a",
			Change: requireAdded,
			New:    &Require{Path: "example.com/a", Version: "v1.0.0"},
		}},
	}}
	if !reflect.DeepEqual(got.Modules, want) {
		t.Fatalf("Compare().Modules = %#v, want %#v", got.Modules, want)
	}
}

func TestCompareModulesSortsByPath(t *testing.T) {
	defs := func(paths ...string) []*model.Definition {
		var result []*model.Definition
		for _, path := range paths {
			d := def(path, "x", false, nil, nil)
			d.Module = &model.Module{Path: path, GoVersion: "1.27.0"}
			result = append(result, d)
		}
		return result
	}

	got := Compare(nil, defs("example.com/c", "example.com/a", "example.com/b"), false, false)

	var paths []string
	for _, change := range got.Modules {
		paths = append(paths, change.Path)
	}
	if want := []string{"example.com/a", "example.com/b", "example.com/c"}; !reflect.DeepEqual(paths, want) {
		t.Errorf("module order = %#v, want %#v", paths, want)
	}
}

func TestCompareModulesReportsNothingForAModelWithoutOne(t *testing.T) {
	// A model extracted from a tree holding no go.mod, or by a version of the
	// tool that did not record one, has nothing to compare and says so.
	defs := []*model.Definition{def("example.com/x", "x", false, nil, nil)}

	got := Compare(defs, defs, false, false)
	if len(got.Modules) != 0 {
		t.Fatalf("Compare().Modules = %#v, want none", got.Modules)
	}
}

func TestRequireSummary(t *testing.T) {
	tests := []struct {
		title  string
		change RequireChange
		want   string
	}{
		{
			title: "added",
			change: RequireChange{
				Path: "example.com/a", Change: requireAdded,
				New: &Require{Path: "example.com/a", Version: "v1.0.0"},
			},
			want: "example.com/a v1.0.0",
		},
		{
			title: "added indirect",
			change: RequireChange{
				Path: "example.com/a", Change: requireAdded,
				New: &Require{Path: "example.com/a", Version: "v1.0.0", Indirect: true},
			},
			want: "example.com/a v1.0.0 // indirect",
		},
		{
			title: "moved to another version",
			change: RequireChange{
				Path: "example.com/a", Change: requireChanged,
				Old: &Require{Path: "example.com/a", Version: "v1.0.0"},
				New: &Require{Path: "example.com/a", Version: "v1.1.0"},
			},
			want: "example.com/a v1.0.0 -> v1.1.0",
		},
		{
			title: "became direct",
			change: RequireChange{
				Path: "example.com/a", Change: requireChanged,
				Old: &Require{Path: "example.com/a", Version: "v1.0.0", Indirect: true},
				New: &Require{Path: "example.com/a", Version: "v1.0.0"},
			},
			want: "example.com/a v1.0.0 (indirect -> direct)",
		},
	}

	for _, test := range tests {
		if got := requireSummary(test.change); got != test.want {
			t.Errorf("%s: requireSummary() = %q, want %q", test.title, got, test.want)
		}
	}
}

func TestReplaceSummary(t *testing.T) {
	tests := []struct {
		title  string
		change ReplaceChange
		want   string
	}{
		{
			title: "added",
			change: ReplaceChange{
				Path: "example.com/a", Change: requireAdded,
				New: &Replace{Path: "example.com/a", NewPath: "example.com/fork/a", NewVersion: "v1.2.0"},
			},
			want: "example.com/a => example.com/fork/a v1.2.0",
		},
		{
			title: "a directory carries no version",
			change: ReplaceChange{
				Path: "example.com/a", Version: "v1.0.0", Change: requireRemoved,
				Old: &Replace{Path: "example.com/a", Version: "v1.0.0", NewPath: "../a"},
			},
			want: "example.com/a v1.0.0 => ../a",
		},
		{
			title: "repointed",
			change: ReplaceChange{
				Path: "example.com/a", Change: requireChanged,
				Old: &Replace{Path: "example.com/a", NewPath: "../a"},
				New: &Replace{Path: "example.com/a", NewPath: "example.com/fork/a", NewVersion: "v1.2.0"},
			},
			want: "example.com/a => ../a -> example.com/fork/a v1.2.0",
		},
	}

	for _, test := range tests {
		if got := replaceSummary(test.change); got != test.want {
			t.Errorf("%s: replaceSummary() = %q, want %q", test.title, got, test.want)
		}
	}
}
