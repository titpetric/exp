package diff

import (
	"reflect"
	"testing"

	"github.com/titpetric/exp/cmd/go-fsck/model"
)

// sym builds the Symbol a package level declaration of example.com/x reports.
func sym(kind, name, signature string) Symbol {
	return Symbol{
		Key:       "example.com/x." + name,
		Package:   "example.com/x",
		Name:      name,
		Kind:      kind,
		Signature: signature,
	}
}

// keys reduces a symbol list to the keys it holds, for the cases where the
// rendering is not what is under test.
func keys(symbols []Symbol) []string {
	result := make([]string, 0, len(symbols))
	for _, symbol := range symbols {
		result = append(result, symbol.Key)
	}
	return result
}

func TestCompare(t *testing.T) {
	tests := []struct {
		title string
		old   []*model.Definition
		new   []*model.Definition
		want  Result
	}{
		{
			title: "no change",
			old: []*model.Definition{def("example.com/x", "x", false,
				model.DeclarationList{{Kind: "func", Name: "Open", Signature: "Open (name string) error"}}, nil)},
			new: []*model.Definition{def("example.com/x", "x", false,
				model.DeclarationList{{Kind: "func", Name: "Open", Signature: "Open (name string) error"}}, nil)},
			want: Result{Removed: []Symbol{}, Added: []Symbol{}, Changed: []Change{}},
		},
		{
			title: "renaming a parameter is not a change",
			old: []*model.Definition{def("example.com/x", "x", false,
				model.DeclarationList{{Kind: "func", Name: "Open", Signature: "Open (name string) error"}}, nil)},
			new: []*model.Definition{def("example.com/x", "x", false,
				model.DeclarationList{{Kind: "func", Name: "Open", Signature: "Open (path string) error"}}, nil)},
			want: Result{Removed: []Symbol{}, Added: []Symbol{}, Changed: []Change{}},
		},
		{
			title: "added symbol is not breaking",
			old: []*model.Definition{def("example.com/x", "x", false, nil,
				model.DeclarationList{{Kind: "type", Name: "Client"}})},
			new: []*model.Definition{def("example.com/x", "x", false, nil,
				model.DeclarationList{{Kind: "type", Names: []string{"Client", "Server"}}})},
			want: Result{Removed: []Symbol{}, Added: []Symbol{sym(kindType, "Server", "")}, Changed: []Change{}},
		},
		{
			title: "removed symbol is breaking",
			old: []*model.Definition{def("example.com/x", "x", false, nil,
				model.DeclarationList{{Kind: "type", Names: []string{"Client", "Server"}}})},
			new: []*model.Definition{def("example.com/x", "x", false, nil,
				model.DeclarationList{{Kind: "type", Name: "Client"}})},
			want: Result{Removed: []Symbol{sym(kindType, "Server", "")}, Added: []Symbol{}, Changed: []Change{}, Breaking: true},
		},
		{
			title: "removed package is breaking",
			old: []*model.Definition{def("example.com/x", "x", false, nil,
				model.DeclarationList{{Kind: "type", Name: "Client"}})},
			new:  []*model.Definition{},
			want: Result{Removed: []Symbol{sym(kindType, "Client", "")}, Added: []Symbol{}, Changed: []Change{}, Breaking: true},
		},
		{
			title: "a method carries its receiver",
			old:   []*model.Definition{def("example.com/x", "x", false, nil, nil)},
			new: []*model.Definition{def("example.com/x", "x", false,
				model.DeclarationList{{Kind: "func", Name: "Close", Receiver: "*Client", Signature: "Close () error"}}, nil)},
			want: Result{
				Removed: []Symbol{},
				Added:   []Symbol{sym(kindFunc, "Client.Close", "func (*Client) Close () error")},
				Changed: []Change{},
			},
		},
		{
			title: "changed signature is breaking",
			old: []*model.Definition{def("example.com/x", "x", false,
				model.DeclarationList{{Kind: "func", Name: "Open", Signature: "Open (name string) error"}}, nil)},
			new: []*model.Definition{def("example.com/x", "x", false,
				model.DeclarationList{{Kind: "func", Name: "Open", Signature: "Open (name string, mode int) error"}}, nil)},
			want: Result{
				Removed: []Symbol{},
				Added:   []Symbol{},
				Changed: []Change{{
					Key:     "example.com/x.Open",
					Package: "example.com/x",
					Name:    "Open",
					Old:     "Open (string) error",
					New:     "Open (string, int) error",
				}},
				Breaking: true,
			},
		},
		{
			title: "moving a declaration between files is not a change",
			old: []*model.Definition{def("example.com/x", "x", false,
				model.DeclarationList{{Kind: "func", Name: "Open", File: "open.go", Line: 3, Signature: "Open ()"}}, nil)},
			new: []*model.Definition{def("example.com/x", "x", false,
				model.DeclarationList{{Kind: "func", Name: "Open", File: "x.go", Line: 91, Signature: "Open ()"}}, nil)},
			want: Result{Removed: []Symbol{}, Added: []Symbol{}, Changed: []Change{}},
		},
	}

	for _, test := range tests {
		got := Compare(test.old, test.new, false)
		if !reflect.DeepEqual(got, test.want) {
			t.Errorf("%s: Compare() = %#v, want %#v", test.title, got, test.want)
		}
	}
}

func TestCompareReportsTheKindOfEachSymbol(t *testing.T) {
	new := []*model.Definition{{
		Package: model.Package{ImportPath: "example.com/x", Package: "x"},
		Types:   model.DeclarationList{{Kind: "struct", Name: "Client"}},
		Consts:  model.DeclarationList{{Kind: "const", Name: "Name"}},
		Vars:    model.DeclarationList{{Kind: "var", Name: "ErrClosed"}},
		Funcs:   model.DeclarationList{{Kind: "func", Name: "Open", Signature: "Open () error"}},
	}}

	var got []string
	for _, symbol := range Compare(nil, new, false).Added {
		got = append(got, symbol.String())
	}
	// Sorted by key, so by name; a struct is declared as one but reads as a
	// type.
	want := []string{"type Client", "var ErrClosed", "const Name", "func Open () error"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Compare().Added = %#v, want %#v", got, want)
	}
}

func TestCompareSortsResults(t *testing.T) {
	old := []*model.Definition{def("example.com/x", "x", false, nil,
		model.DeclarationList{{Kind: "type", Names: []string{"C", "A", "B"}}})}

	got := Compare(old, []*model.Definition{}, false)
	want := []string{"example.com/x.A", "example.com/x.B", "example.com/x.C"}
	if !reflect.DeepEqual(keys(got.Removed), want) {
		t.Fatalf("Compare().Removed = %#v, want %#v", keys(got.Removed), want)
	}
}
