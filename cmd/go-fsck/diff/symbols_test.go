package diff

import (
	"reflect"
	"sort"
	"testing"

	"github.com/titpetric/exp/cmd/go-fsck/model"
)

// def builds a definition holding a single declaration list, the shape the
// tests compare against.
func def(importPath, pkg string, testPackage bool, funcs, types model.DeclarationList) *model.Definition {
	d := &model.Definition{Funcs: funcs, Types: types}
	d.ImportPath = importPath
	d.Package.Package = pkg
	d.TestPackage = testPackage
	return d
}

func keysOf(t *testing.T, defs []*model.Definition, includeInternal bool) []string {
	t.Helper()
	got := symbols(defs, includeInternal)
	keys := make([]string, 0, len(got))
	for key := range got {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func TestSymbolsFiltersUnexportedNamesIndividually(t *testing.T) {
	defs := []*model.Definition{
		def("example.com/x", "x", false, nil, model.DeclarationList{
			// A grouped declaration reports itself as exported when any one
			// of its names is, so the names have to be filtered on their own.
			{Kind: "const", Names: []string{"Public", "private"}},
			{Kind: "type", Name: "Client"},
			{Kind: "type", Name: "hidden"},
		}),
	}

	want := []string{"example.com/x.Client", "example.com/x.Public"}
	if got := keysOf(t, defs, false); !reflect.DeepEqual(got, want) {
		t.Fatalf("symbols() = %#v, want %#v", got, want)
	}
}

func TestSymbolsDropsMethodsOnUnexportedReceivers(t *testing.T) {
	defs := []*model.Definition{
		def("example.com/x", "x", false, model.DeclarationList{
			{Kind: "func", Name: "Close", Receiver: "*Client", Signature: "Close () error"},
			{Kind: "func", Name: "Close", Receiver: "*collector", Signature: "Close () error"},
			{Kind: "func", Name: "Open", Signature: "Open (name string) error"},
		}, nil),
	}

	want := []string{"example.com/x.Client.Close", "example.com/x.Open"}
	if got := keysOf(t, defs, false); !reflect.DeepEqual(got, want) {
		t.Fatalf("symbols() = %#v, want %#v", got, want)
	}
}

func TestSymbolsSkipsTestPackagesAndCommands(t *testing.T) {
	defs := []*model.Definition{
		def("example.com/x", "x_test", true, nil, model.DeclarationList{{Kind: "type", Name: "Fixture"}}),
		def("example.com/x/cmd/x", "main", false, nil, model.DeclarationList{{Kind: "type", Name: "Config"}}),
		def("example.com/x", "x", false, nil, model.DeclarationList{{Kind: "type", Name: "Client"}}),
	}

	want := []string{"example.com/x.Client"}
	if got := keysOf(t, defs, false); !reflect.DeepEqual(got, want) {
		t.Fatalf("symbols() = %#v, want %#v", got, want)
	}
}

func TestSymbolsInternalPackages(t *testing.T) {
	defs := []*model.Definition{
		def("example.com/x/internal/y", "y", false, nil, model.DeclarationList{{Kind: "type", Name: "Hidden"}}),
		def("example.com/x/internally", "internally", false, nil, model.DeclarationList{{Kind: "type", Name: "Shown"}}),
	}

	want := []string{"example.com/x/internally.Shown"}
	if got := keysOf(t, defs, false); !reflect.DeepEqual(got, want) {
		t.Fatalf("symbols() = %#v, want %#v", got, want)
	}

	want = []string{"example.com/x/internal/y.Hidden", "example.com/x/internally.Shown"}
	if got := keysOf(t, defs, true); !reflect.DeepEqual(got, want) {
		t.Fatalf("symbols(includeInternal) = %#v, want %#v", got, want)
	}
}

func TestNormalizeSignature(t *testing.T) {
	tests := []struct {
		signature string
		want      string
	}{
		{"Open ()", "Open ()"},
		{"Open () error", "Open () error"},
		{"Open (name string) error", "Open (string) error"},
		{"Open (name,mode string) error", "Open (string) error"},
		{"Open (name string, mode int) error", "Open (string, int) error"},
		{"Open (_ string) error", "Open (string) error"},
		{"Open (string) error", "Open (string) error"},
		{"Open (names ...string) error", "Open (...string) error"},
		{"Open (...string) error", "Open (...string) error"},
		{"Open (out chan int) error", "Open (chan int) error"},
		{"Open (chan int) error", "Open (chan int) error"},
		{"Open (v struct { A int }) error", "Open (struct { A int }) error"},
		{"Open (m map[string]int, s []byte) (int, error)", "Open (map[string]int, []byte) (int, error)"},
		// A comma inside a parameter type does not separate parameters.
		{"Walk (fn func(a, b int) error) error", "Walk (func(a, b int) error) error"},
		{"Walk (root string, fn func(a, b int) error) error", "Walk (string, func(a, b int) error) error"},
		// Nothing to work with is returned as it stands.
		{"Open", "Open"},
		{"Open (name string", "Open (name string"},
	}

	for _, test := range tests {
		if got := normalizeSignature(test.signature); got != test.want {
			t.Errorf("normalizeSignature(%q) = %q, want %q", test.signature, got, test.want)
		}
	}
}

func TestIsNameList(t *testing.T) {
	tests := map[string]bool{
		"a":      true,
		"a,b":    true,
		"_":      true,
		"":       false,
		"chan":   false,
		"func":   false,
		"struct": false,
		"map":    false,
		"a,":     false,
		"[]byte": false,
	}

	for input, want := range tests {
		if got := isNameList(input); got != want {
			t.Errorf("isNameList(%q) = %v, want %v", input, got, want)
		}
	}
}

func TestSourceWithoutDoc(t *testing.T) {
	tests := []struct {
		title  string
		source string
		want   string
	}{
		{
			title:  "no source to work with",
			source: "",
			want:   "",
		},
		{
			title:  "a declaration with no doc comment",
			source: "type Tag struct {\n\tName string\n}",
			want:   "type Tag struct {\n\tName string\n}",
		},
		{
			title:  "the doc comment above it is dropped",
			source: "// Tag is a release tag.\n// It is what a version is read from.\ntype Tag struct {\n\tName string\n}",
			want:   "type Tag struct {\n\tName string\n}",
		},
		{
			title:  "a comment inside the body stays where it is",
			source: "// Tag is a release tag.\ntype Tag struct {\n\t// Name is the tag itself.\n\tName string\n}",
			want:   "type Tag struct {\n\t// Name is the tag itself.\n\tName string\n}",
		},
		{
			title:  "nothing but a comment leaves nothing",
			source: "// Tag is a release tag.",
			want:   "",
		},
	}

	for _, test := range tests {
		if got := sourceWithoutDoc(test.source); got != test.want {
			t.Errorf("%s: sourceWithoutDoc(%q) = %q, want %q", test.title, test.source, got, test.want)
		}
	}
}

func TestSymbolsCarriesTheBodyOfATypeOnly(t *testing.T) {
	defs := []*model.Definition{{
		Package: model.Package{ImportPath: "example.com/x", Package: "x"},
		Types: model.DeclarationList{{
			Kind:   "type",
			Name:   "Tag",
			Source: "// Tag is a release tag.\ntype Tag struct {\n\tMajor, Minor, Patch uint64\n}",
		}},
		Funcs: model.DeclarationList{{
			Kind:      "func",
			Name:      "Parse",
			Signature: "Parse (s string) (Tag, bool)",
			Source:    "// Parse reads a tag.\nfunc Parse(s string) (Tag, bool) {\n\treturn Tag{}, false\n}",
		}},
	}}

	got := symbols(defs, false)

	// The body keeps every grouped field name, which is what reading it from
	// the source rather than the decomposed fields is for.
	tag := got["example.com/x.Tag"].symbol
	want := "type Tag struct {\n\tMajor, Minor, Patch uint64\n}"
	if tag.Definition != want {
		t.Errorf("Tag.Definition = %q, want %q", tag.Definition, want)
	}

	// A func is named by its signature; its body is no use to a reader.
	fn := got["example.com/x.Parse"].symbol
	if fn.Definition != "" {
		t.Errorf("Parse.Definition = %q, want it empty", fn.Definition)
	}
	if fn.Signature != "func Parse (s string) (Tag, bool)" {
		t.Errorf("Parse.Signature = %q", fn.Signature)
	}
}
