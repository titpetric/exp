package diff

import (
	"encoding/json"
	"fmt"
	"sort"

	"github.com/titpetric/exp/cmd/go-fsck/model"
	"github.com/titpetric/exp/cmd/go-fsck/model/loader"
)

// Symbol is one exported declaration of a package.
type Symbol struct {
	// Key identifies the declaration across revisions, as the import path
	// followed by the receiver type and name.
	Key string `json:"key"`

	// Package is the import path holding the declaration.
	Package string `json:"package"`

	// Name is the declared name, qualified with its receiver type when it
	// has one, as in "Manager.Start".
	Name string `json:"name"`

	// Kind is type, const, var or func.
	Kind string `json:"kind"`

	// Signature is a func as it is declared, and is empty for every other
	// kind, which is named by Kind and Name alone.
	Signature string `json:"signature,omitempty"`

	// Definition is a type as it is declared, without its doc comment. It is
	// empty for every other kind, and for a type whenever the model was
	// extracted without sources.
	Definition string `json:"definition,omitempty"`

	// Underlying is the shape a type is declared with: "struct", "interface",
	// or the type it is defined as. It is empty for every other kind.
	Underlying string `json:"underlying,omitempty"`

	// Fields are the exported fields of a struct, or the methods of an
	// interface, in name order. They are empty for every other kind, and for a
	// type holding none.
	Fields []Field `json:"fields,omitempty"`
}

// Field is one exported field of a struct, or one method of an interface.
type Field struct {
	// Name is the field name, and for an embedded field the name it is
	// reached by.
	Name string `json:"name"`

	// Type is the field type, and for an interface method its signature.
	Type string `json:"type"`

	// Tag is the struct tag, unmodified, and is empty when the field carries
	// none.
	Tag string `json:"tag,omitempty"`

	// Embedded reports a field declared without a name of its own, which
	// promotes the method set of the type it embeds.
	Embedded bool `json:"embedded,omitempty"`
}

// The changes a field of a type undergoes between two revisions.
const (
	fieldAdded   = "added"
	fieldChanged = "changed"
	fieldRemoved = "removed"
)

// FieldChange is one exported field a release adds to, drops from, or reshapes
// on a type both revisions carry.
type FieldChange struct {
	// Name is the field the change is to.
	Name string `json:"name"`

	// Change is added, changed or removed.
	Change string `json:"change"`

	// Old and New are the field before and after. Old is absent for a field
	// that was added, and New for one that was removed.
	Old *Field `json:"old,omitempty"`
	New *Field `json:"new,omitempty"`
}

// TypeChange is a type both revisions carry whose exported fields moved. It is
// the data model side of a release: a struct field and an interface method are
// promises to a consumer as much as a func signature is.
type TypeChange struct {
	Key     string `json:"key"`
	Package string `json:"package"`
	Name    string `json:"name"`

	// Underlying is the shape the type is declared with, which is the same in
	// both revisions: a type that changed shape is a changed symbol instead.
	Underlying string `json:"underlying"`

	// Fields are the changes, in name order.
	Fields []FieldChange `json:"fields"`

	// Breaking reports whether the change costs a consumer something. Taking a
	// field away or reshaping one does; adding one to a struct does not, while
	// adding a method to an interface stops every implementor compiling.
	Breaking bool `json:"breaking"`
}

// String renders the declaration the way it reads in source.
func (s Symbol) String() string {
	if s.Signature != "" {
		return s.Signature
	}
	return s.Kind + " " + s.Name
}

// String renders one field the way it is declared, which for an embedded field
// is the type it embeds, pointer and package qualifier included. The name an
// embed is reached by is the type with both stripped, so naming it instead
// says less than the source does.
func (f Field) String() string {
	if f.Embedded {
		return "embeds " + f.Type
	}
	return f.Name + " " + f.Type
}

// Field returns the side of the change to report, which is the field as it
// stands after the release, or as it stood when the release dropped it.
func (c FieldChange) Field() Field {
	switch {
	case c.New != nil:
		return *c.New
	case c.Old != nil:
		return *c.Old
	}
	return Field{Name: c.Name}
}

// Label returns how the change reads under the type it is a field of: an
// embedded field says what it embeds, and every other field is reached by
// name.
func (c FieldChange) Label(key string) string {
	if field := c.Field(); field.Embedded {
		return key + " " + field.String()
	}
	return key + "." + c.Name
}

// Change is an exported symbol whose signature moved between two revisions.
type Change struct {
	Key     string `json:"key"`
	Package string `json:"package"`
	Name    string `json:"name"`

	// Old and New are the signature before and after, with parameter names
	// removed, which is what they were compared on.
	Old string `json:"old"`
	New string `json:"new"`
}

// Result is the exported API difference between two revisions of a module.
type Result struct {
	// Removed holds the symbols the older revision had and the newer one
	// does not.
	Removed []Symbol `json:"removed"`

	// Added holds the symbols the newer revision gained.
	Added []Symbol `json:"added"`

	// Changed holds the symbols both revisions have under a signature that
	// moved.
	Changed []Change `json:"changed"`

	// Types holds the types both revisions have whose exported fields moved,
	// which is the data model the release changes.
	Types []TypeChange `json:"types"`

	// Modules holds the go.mod changes, which is what the release changes
	// about depending on the module rather than about calling it.
	Modules []ModuleChange `json:"modules"`

	// Breaking reports whether the difference takes API away, which is a
	// removed symbol, a changed signature, or a data model change that costs a
	// consumer something. Added symbols are not breaking, and neither are
	// module changes: a dependency is not API.
	Breaking bool `json:"breaking"`
}

// Compare returns the difference between two sets of definitions. Symbols are
// keyed by import path, receiver type and name, so a declaration moving to
// another file, or its group gaining a sibling, is not a difference.
//
// The go.mod of each module is compared alongside the symbols, since a release
// changes what it costs to depend on a module as much as it changes what the
// module offers. Indirect requirements are left out unless includeIndirect
// asks for them.
func Compare(oldDefs, newDefs []*model.Definition, includeInternal, includeIndirect bool) Result {
	var (
		old = symbols(oldDefs, includeInternal)
		cur = symbols(newDefs, includeInternal)
	)

	result := Result{
		Removed: []Symbol{},
		Added:   []Symbol{},
		Changed: []Change{},
		Types:   []TypeChange{},
		Modules: compareModules(modules(oldDefs), modules(newDefs), includeIndirect),
	}

	for key, was := range old {
		is, ok := cur[key]
		switch {
		case !ok:
			result.Removed = append(result.Removed, was.symbol)
		case is.normalized != was.normalized:
			result.Changed = append(result.Changed, Change{
				Key:     key,
				Package: is.symbol.Package,
				Name:    is.symbol.Name,
				Old:     was.normalized,
				New:     is.normalized,
			})
		default:
			// The shape held, so what is left to compare is what the shape is
			// made of.
			if change, ok := compareFields(key, was, is); ok {
				result.Types = append(result.Types, change)
			}
		}
	}
	for key, is := range cur {
		if _, ok := old[key]; !ok {
			result.Added = append(result.Added, is.symbol)
		}
	}

	// The maps are walked in whatever order the runtime hands out, and the
	// declaration order of the input is not dependable either, so the lists
	// are sorted to make the output of two identical runs identical.
	sortSymbols(result.Removed)
	sortSymbols(result.Added)
	sort.Slice(result.Changed, func(i, j int) bool { return result.Changed[i].Key < result.Changed[j].Key })
	sort.Slice(result.Types, func(i, j int) bool { return result.Types[i].Key < result.Types[j].Key })

	result.Breaking = len(result.Removed) > 0 || len(result.Changed) > 0
	for _, change := range result.Types {
		result.Breaking = result.Breaking || change.Breaking
	}
	return result
}

// compareFields reports how the exported fields of one type moved between two
// revisions, and whether they moved at all.
//
// Both entries describe the same type under the same shape, so the fields are
// compared by name: a field the newer revision does not carry was removed, one
// the older did not carry was added, and one both carry under another type or
// another tag was reshaped.
//
// A tag is compared along with the type. It is what a document decodes through,
// so renaming a json or yaml key breaks every document already written even
// though the code that reads it still compiles.
func compareFields(key string, was, is entry) (TypeChange, bool) {
	change := TypeChange{
		Key:        key,
		Package:    is.symbol.Package,
		Name:       is.symbol.Name,
		Underlying: is.symbol.Underlying,
		Fields:     []FieldChange{},
	}

	// An interface promises a method set in both directions: a consumer calls
	// what it names, and an implementor satisfies it. Every move breaks one of
	// the two, where a struct only breaks on what it takes away.
	isInterface := is.symbol.Underlying == underlyingInterface

	for name, before := range was.fields {
		after, ok := is.fields[name]
		switch {
		case !ok:
			change.Fields = append(change.Fields, FieldChange{
				Name: name, Change: fieldRemoved, Old: fieldPtr(before),
			})
			change.Breaking = true
		case fieldMoved(before, after, isInterface):
			change.Fields = append(change.Fields, FieldChange{
				Name: name, Change: fieldChanged, Old: fieldPtr(before), New: fieldPtr(after),
			})
			change.Breaking = true
		}
	}
	for name, after := range is.fields {
		if _, ok := was.fields[name]; !ok {
			change.Fields = append(change.Fields, FieldChange{
				Name: name, Change: fieldAdded, New: fieldPtr(after),
			})
			change.Breaking = change.Breaking || isInterface
		}
	}

	if len(change.Fields) == 0 {
		return TypeChange{}, false
	}
	sort.Slice(change.Fields, func(i, j int) bool { return change.Fields[i].Name < change.Fields[j].Name })
	return change, true
}

// fieldMoved reports whether a field both revisions carry is no longer the
// same one. An interface method is compared the way a func is, so renaming one
// of its parameters is not a change while retyping one is.
func fieldMoved(before, after Field, isInterface bool) bool {
	if before.Tag != after.Tag || before.Embedded != after.Embedded {
		return true
	}
	if isInterface {
		return normalizeSignature(before.Type) != normalizeSignature(after.Type)
	}
	return before.Type != after.Type
}

// fieldPtr returns a field to report, which the result holds by pointer so the
// side a change has no field on is left out of the json.
func fieldPtr(field Field) *Field {
	return &field
}

func sortSymbols(symbols []Symbol) {
	sort.Slice(symbols, func(i, j int) bool { return symbols[i].Key < symbols[j].Key })
}

func diff(cfg *options) error {
	oldDefs, err := loader.ReadFile(cfg.oldFile)
	if err != nil {
		return fmt.Errorf("read %s: %w", cfg.oldFile, err)
	}
	newDefs, err := loader.ReadFile(cfg.newFile)
	if err != nil {
		return fmt.Errorf("read %s: %w", cfg.newFile, err)
	}

	result := Compare(oldDefs, newDefs, cfg.includeInternal, cfg.includeIndirect)

	if cfg.json {
		b, err := json.Marshal(result)
		if err != nil {
			return err
		}
		fmt.Println(string(b))
		return nil
	}

	for _, symbol := range result.Removed {
		fmt.Printf("- %s\n", symbol.Key)
		if cfg.verbose {
			fmt.Printf("    %s\n", symbol)
		}
	}
	for _, change := range result.Changed {
		fmt.Printf("~ %s\n", change.Key)
		if cfg.verbose {
			fmt.Printf("    %s\n    %s\n", change.Old, change.New)
		}
	}
	for _, symbol := range result.Added {
		fmt.Printf("+ %s\n", symbol.Key)
		if cfg.verbose {
			fmt.Printf("    %s\n", symbol)
		}
	}
	for _, change := range result.Types {
		for _, field := range change.Fields {
			fmt.Printf("%s %s\n", fieldMarker(field.Change), field.Label(change.Key))
			if cfg.verbose {
				if field.Old != nil {
					fmt.Printf("    %s\n", field.Old.Type)
				}
				if field.New != nil {
					fmt.Printf("    %s\n", field.New.Type)
				}
			}
		}
	}

	for _, change := range result.Modules {
		if cfg.verbose {
			fmt.Printf("module %s\n", change.Path)
		}
		if change.Go != nil {
			fmt.Printf("~ go %s -> %s\n", change.Go.Old, change.Go.New)
		}
		if change.Toolchain != nil {
			fmt.Printf("~ toolchain %s -> %s\n", change.Toolchain.Old, change.Toolchain.New)
		}
		for _, require := range change.Requires {
			fmt.Printf("%s require %s\n", fieldMarker(require.Change), requireSummary(require))
		}
		for _, replace := range change.Replaces {
			fmt.Printf("%s replace %s\n", fieldMarker(replace.Change), replaceSummary(replace))
		}
	}

	summary := fmt.Sprintf("%d removed, %d changed, %d added, %d fields, %d requires", len(result.Removed), len(result.Changed), len(result.Added), countFields(result.Types), countRequires(result.Modules))
	if result.Breaking {
		summary += ", breaking"
	}
	fmt.Println(summary)
	return nil
}

// fieldMarker returns the marker a field change is listed under, which is the
// one the symbol it belongs to would be listed under.
func fieldMarker(change string) string {
	switch change {
	case fieldAdded:
		return "+"
	case fieldRemoved:
		return "-"
	}
	return "~"
}

// countFields returns how many fields moved across every type reported.
func countFields(types []TypeChange) int {
	count := 0
	for _, change := range types {
		count += len(change.Fields)
	}
	return count
}
