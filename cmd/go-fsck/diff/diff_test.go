package diff

import (
	"reflect"
	"testing"

	"github.com/titpetric/tools/splint/model"
)

// sym builds the Symbol a package level declaration of example.com/x reports.
func sym(kind, name, signature string) Symbol {
	return Symbol{
		Key:       "example.com/x." + name,
		Package:   "example.com/x",
		Name:      name,
		Kind:      kind,
		Exported:  true,
		Signature: signature,
	}
}

// symType builds the Symbol a type declaration of example.com/x reports, under
// the shape it is declared with.
func symType(name, underlying string, fields ...Field) Symbol {
	symbol := sym(kindType, name, "")
	symbol.Underlying = underlying
	symbol.Fields = fields
	return symbol
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
			want: Result{Removed: []Symbol{}, Added: []Symbol{}, Changed: []Change{}, Types: []TypeChange{}, Modules: []ModuleChange{}},
		},
		{
			title: "renaming a parameter is not a change",
			old: []*model.Definition{def("example.com/x", "x", false,
				model.DeclarationList{{Kind: "func", Name: "Open", Signature: "Open (name string) error"}}, nil)},
			new: []*model.Definition{def("example.com/x", "x", false,
				model.DeclarationList{{Kind: "func", Name: "Open", Signature: "Open (path string) error"}}, nil)},
			want: Result{Removed: []Symbol{}, Added: []Symbol{}, Changed: []Change{}, Types: []TypeChange{}, Modules: []ModuleChange{}},
		},
		{
			title: "added symbol is not breaking",
			old: []*model.Definition{def("example.com/x", "x", false, nil,
				model.DeclarationList{{Kind: "type", Name: "Client"}})},
			new: []*model.Definition{def("example.com/x", "x", false, nil,
				model.DeclarationList{{Kind: "type", Names: []string{"Client", "Server"}}})},
			want: Result{Removed: []Symbol{}, Added: []Symbol{symType("Server", underlyingStruct)}, Changed: []Change{}, Types: []TypeChange{}, Modules: []ModuleChange{}},
		},
		{
			title: "removed symbol is breaking",
			old: []*model.Definition{def("example.com/x", "x", false, nil,
				model.DeclarationList{{Kind: "type", Names: []string{"Client", "Server"}}})},
			new: []*model.Definition{def("example.com/x", "x", false, nil,
				model.DeclarationList{{Kind: "type", Name: "Client"}})},
			want: Result{Removed: []Symbol{symType("Server", underlyingStruct)}, Added: []Symbol{}, Changed: []Change{}, Types: []TypeChange{}, Modules: []ModuleChange{}, Breaking: true},
		},
		{
			title: "removed package is breaking",
			old: []*model.Definition{def("example.com/x", "x", false, nil,
				model.DeclarationList{{Kind: "type", Name: "Client"}})},
			new:  []*model.Definition{},
			want: Result{Removed: []Symbol{symType("Client", underlyingStruct)}, Added: []Symbol{}, Changed: []Change{}, Types: []TypeChange{}, Modules: []ModuleChange{}, Breaking: true},
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
				Types:   []TypeChange{},
				Modules: []ModuleChange{},
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
					Key:      "example.com/x.Open",
					Package:  "example.com/x",
					Name:     "Open",
					Exported: true,
					Old:      "Open (string) error",
					New:      "Open (string, int) error",
				}},
				Types:    []TypeChange{},
				Modules:  []ModuleChange{},
				Breaking: true,
			},
		},
		{
			title: "moving a declaration between files is not a change",
			old: []*model.Definition{def("example.com/x", "x", false,
				model.DeclarationList{{Kind: "func", Name: "Open", File: "open.go", Line: 3, Signature: "Open ()"}}, nil)},
			new: []*model.Definition{def("example.com/x", "x", false,
				model.DeclarationList{{Kind: "func", Name: "Open", File: "x.go", Line: 91, Signature: "Open ()"}}, nil)},
			want: Result{Removed: []Symbol{}, Added: []Symbol{}, Changed: []Change{}, Types: []TypeChange{}, Modules: []ModuleChange{}},
		},
	}

	for _, test := range tests {
		got := Compare(test.old, test.new, false, false)
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
	for _, symbol := range Compare(nil, new, false, false).Added {
		got = append(got, symbol.String())
	}
	// Sorted by key, so by name; a struct is declared as one but reads as a
	// type.
	want := []string{"type Client struct", "var ErrClosed", "const Name", "func Open () error"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Compare().Added = %#v, want %#v", got, want)
	}
}

func TestCompareSortsResults(t *testing.T) {
	old := []*model.Definition{def("example.com/x", "x", false, nil,
		model.DeclarationList{{Kind: "type", Names: []string{"C", "A", "B"}}})}

	got := Compare(old, []*model.Definition{}, false, false)
	want := []string{"example.com/x.A", "example.com/x.B", "example.com/x.C"}
	if !reflect.DeepEqual(keys(got.Removed), want) {
		t.Fatalf("Compare().Removed = %#v, want %#v", keys(got.Removed), want)
	}
}

// typeDefs builds a definition holding one type declaration, which is what the
// data model comparison reads.
func typeDefs(name, underlying string, fields ...*model.Field) []*model.Definition {
	decl := &model.Declaration{Kind: "type", Name: name, Fields: fields}
	if underlying != underlyingStruct {
		decl.Type = underlying
	}
	return []*model.Definition{def("example.com/x", "x", false, nil, model.DeclarationList{decl})}
}

func TestCompareReportsDataModelChanges(t *testing.T) {
	tests := []struct {
		title string
		old   []*model.Definition
		new   []*model.Definition
		want  []FieldChange
		// breaking is the verdict on the type, which differs between a struct
		// and an interface for an added field.
		breaking bool
	}{
		{
			title: "a field added to a struct is not breaking",
			old:   typeDefs("Config", underlyingStruct, &model.Field{Name: "Addr", Type: "string"}),
			new: typeDefs("Config", underlyingStruct,
				&model.Field{Name: "Addr", Type: "string"},
				&model.Field{Name: "Port", Type: "int"}),
			want: []FieldChange{{
				Name: "Port", Change: fieldAdded, New: &Field{Name: "Port", Type: "int"},
			}},
		},
		{
			title: "a field taken off a struct is breaking",
			old: typeDefs("Config", underlyingStruct,
				&model.Field{Name: "Addr", Type: "string"},
				&model.Field{Name: "Port", Type: "int"}),
			new:      typeDefs("Config", underlyingStruct, &model.Field{Name: "Addr", Type: "string"}),
			want:     []FieldChange{{Name: "Port", Change: fieldRemoved, Old: &Field{Name: "Port", Type: "int"}}},
			breaking: true,
		},
		{
			title: "a retyped field is breaking",
			old:   typeDefs("Config", underlyingStruct, &model.Field{Name: "Addr", Type: "string"}),
			new:   typeDefs("Config", underlyingStruct, &model.Field{Name: "Addr", Type: "[]string"}),
			want: []FieldChange{{
				Name: "Addr", Change: fieldChanged,
				Old: &Field{Name: "Addr", Type: "string"},
				New: &Field{Name: "Addr", Type: "[]string"},
			}},
			breaking: true,
		},
		{
			title: "a renamed tag key is breaking, since a stored document decodes through it",
			old:   typeDefs("Config", underlyingStruct, &model.Field{Name: "Addr", Type: "string", Tag: `yaml:"addr"`}),
			new:   typeDefs("Config", underlyingStruct, &model.Field{Name: "Addr", Type: "string", Tag: `yaml:"address"`}),
			want: []FieldChange{{
				Name: "Addr", Change: fieldChanged,
				Old: &Field{Name: "Addr", Type: "string", Tag: `yaml:"addr"`},
				New: &Field{Name: "Addr", Type: "string", Tag: `yaml:"address"`},
			}},
			breaking: true,
		},
		{
			title: "an unexported field is nobody's promise",
			old:   typeDefs("Config", underlyingStruct, &model.Field{Name: "secret", Type: "string"}),
			new:   typeDefs("Config", underlyingStruct, &model.Field{Name: "secret", Type: "int"}),
			want:  nil,
		},
		{
			title: "a method added to an interface is breaking, since it stops every implementor compiling",
			old:   typeDefs("Store", underlyingInterface, &model.Field{Name: "Get", Type: "Get (key string) error"}),
			new: typeDefs("Store", underlyingInterface,
				&model.Field{Name: "Get", Type: "Get (key string) error"},
				&model.Field{Name: "Put", Type: "Put (key string) error"}),
			want: []FieldChange{{
				Name: "Put", Change: fieldAdded, New: &Field{Name: "Put", Type: "Put (key string) error"},
			}},
			breaking: true,
		},
		{
			title: "renaming the parameter of an interface method is not a change",
			old:   typeDefs("Store", underlyingInterface, &model.Field{Name: "Get", Type: "Get (key string) error"}),
			new:   typeDefs("Store", underlyingInterface, &model.Field{Name: "Get", Type: "Get (name string) error"}),
			want:  nil,
		},
		{
			title: "retyping the parameter of an interface method is a change",
			old:   typeDefs("Store", underlyingInterface, &model.Field{Name: "Get", Type: "Get (key string) error"}),
			new:   typeDefs("Store", underlyingInterface, &model.Field{Name: "Get", Type: "Get (key []byte) error"}),
			want: []FieldChange{{
				Name: "Get", Change: fieldChanged,
				Old: &Field{Name: "Get", Type: "Get (key string) error"},
				New: &Field{Name: "Get", Type: "Get (key []byte) error"},
			}},
			breaking: true,
		},
	}

	for _, test := range tests {
		got := Compare(test.old, test.new, false, false)
		if len(test.want) == 0 {
			if len(got.Types) != 0 {
				t.Errorf("%s: Compare().Types = %#v, want none", test.title, got.Types)
			}
			if got.Breaking {
				t.Errorf("%s: Compare() called it breaking", test.title)
			}
			continue
		}

		if len(got.Types) != 1 {
			t.Errorf("%s: Compare().Types = %#v, want one type", test.title, got.Types)
			continue
		}
		if !reflect.DeepEqual(got.Types[0].Fields, test.want) {
			t.Errorf("%s: fields = %#v, want %#v", test.title, got.Types[0].Fields, test.want)
		}
		if got.Types[0].Breaking != test.breaking || got.Breaking != test.breaking {
			t.Errorf("%s: breaking = %v/%v, want %v", test.title, got.Types[0].Breaking, got.Breaking, test.breaking)
		}
	}
}

func TestCompareReportsAChangedUnderlyingTypeAsAChangedSymbol(t *testing.T) {
	got := Compare(typeDefs("ID", "string"), typeDefs("ID", "int"), false, false)

	want := []Change{{
		Key:      "example.com/x.ID",
		Package:  "example.com/x",
		Name:     "ID",
		Exported: true,
		Old:      "type ID string",
		New:      "type ID int",
	}}
	if !reflect.DeepEqual(got.Changed, want) {
		t.Fatalf("Compare().Changed = %#v, want %#v", got.Changed, want)
	}
	if len(got.Types) != 0 {
		t.Errorf("Compare().Types = %#v, want none: the shape changed, not its fields", got.Types)
	}
	if !got.Breaking {
		t.Error("Compare() did not call a changed underlying type breaking")
	}
}

func TestCompareCarriesTheShapeOfAnAddedType(t *testing.T) {
	got := Compare(nil, typeDefs("Config", underlyingStruct,
		&model.Field{Name: "Addr", Type: "string", Tag: `yaml:"addr"`},
		&model.Field{Name: "secret", Type: "string"}), false, false)

	want := []Symbol{symType("Config", underlyingStruct, Field{Name: "Addr", Type: "string", Tag: `yaml:"addr"`})}
	if !reflect.DeepEqual(got.Added, want) {
		t.Fatalf("Compare().Added = %#v, want %#v", got.Added, want)
	}
}

func TestCompareSortsDataModelChanges(t *testing.T) {
	old := typeDefs("Config", underlyingStruct,
		&model.Field{Name: "C", Type: "int"},
		&model.Field{Name: "A", Type: "int"},
		&model.Field{Name: "B", Type: "int"})
	got := Compare(old, typeDefs("Config", underlyingStruct), false, false)

	if len(got.Types) != 1 {
		t.Fatalf("Compare().Types = %#v, want one type", got.Types)
	}
	var names []string
	for _, field := range got.Types[0].Fields {
		names = append(names, field.Name)
	}
	if want := []string{"A", "B", "C"}; !reflect.DeepEqual(names, want) {
		t.Errorf("field order = %#v, want %#v", names, want)
	}
}

func TestFieldStringReadsAsDeclared(t *testing.T) {
	tests := []struct {
		field Field
		want  string
	}{
		{Field{Name: "Addr", Type: "string"}, "Addr string"},
		{Field{Name: "Storage", Type: "*Storage", Embedded: true}, "embeds *Storage"},
		{Field{Name: "Module", Type: "platform.Module", Embedded: true}, "embeds platform.Module"},
	}
	for _, test := range tests {
		if got := test.field.String(); got != test.want {
			t.Errorf("Field%#v.String() = %q, want %q", test.field, got, test.want)
		}
	}
}

func TestFieldChangeLabelNamesTheEmbeddedType(t *testing.T) {
	added := Field{Name: "Storage", Type: "*Storage", Embedded: true}
	change := FieldChange{Name: "Storage", Change: fieldAdded, New: &added}
	if got, want := change.Label("Disk"), "Disk embeds *Storage"; got != want {
		t.Errorf("Label() = %q, want %q", got, want)
	}

	// A removed field is reported from the side that still has it.
	removed := Field{Name: "Addr", Type: "string"}
	change = FieldChange{Name: "Addr", Change: fieldRemoved, Old: &removed}
	if got, want := change.Label("Disk"), "Disk.Addr"; got != want {
		t.Errorf("Label() = %q, want %q", got, want)
	}
}

func TestCompareWithIncludesUnexported(t *testing.T) {
	defs := []*model.Definition{def("example.com/x", "x", false, model.DeclarationList{
		{Names: []string{"Open"}, Signature: "Open() error"},
		{Names: []string{"parse"}, Signature: "parse() error"},
	}, nil)}

	// The default comparison is the API, so the unexported func is not in it.
	api := CompareWith(Options{}, nil, defs)
	if got := keys(api.Added); len(got) != 1 || got[0] != "example.com/x.Open" {
		t.Fatalf("Compare().Added = %v, want the exported func alone", got)
	}

	full := CompareWith(Options{IncludeUnexported: true}, nil, defs)
	if got := keys(full.Added); len(got) != 2 {
		t.Fatalf("Compare().Added = %v, want both funcs", got)
	}
	for _, symbol := range full.Added {
		if want := symbol.Name == "Open"; symbol.Exported != want {
			t.Errorf("%s reports Exported=%v", symbol.Name, symbol.Exported)
		}
	}
}

func TestCompareWithUnexportedIsNeverBreaking(t *testing.T) {
	before := []*model.Definition{def("example.com/x", "x", false,
		model.DeclarationList{{Names: []string{"parse"}, Signature: "parse() error"}}, nil)}

	// Dropping an unexported func is a refactor, not a release: nothing
	// outside the module compiled against it.
	got := CompareWith(Options{IncludeUnexported: true}, before, nil)
	if len(got.Removed) != 1 {
		t.Fatalf("Compare().Removed = %v, want the unexported func", keys(got.Removed))
	}
	if got.Breaking {
		t.Error("dropping an unexported func was called breaking")
	}
}

func TestCompareWithInternalPackagesAreNotExported(t *testing.T) {
	defs := []*model.Definition{def("example.com/x/internal", "internal", false,
		model.DeclarationList{{Names: []string{"Broker"}, Signature: "Broker() error"}}, nil)}

	got := CompareWith(Options{IncludeInternal: true}, nil, defs)
	if len(got.Added) != 1 {
		t.Fatalf("Compare().Added = %v, want the internal func", keys(got.Added))
	}
	if got.Added[0].Exported {
		t.Error("an exported name in an internal package reported as API")
	}
}
