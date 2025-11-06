package model

import "strings"

// Definition holds all symbols defined for a Package.
type Definition struct {
	Package

	Doc string `json:",omitempty"`

	Imports   StringSet `json:",omitempty"`
	InitCount int       `json:",omitempty"`

	Types  DeclarationList `json:",omitempty"`
	Consts DeclarationList `json:",omitempty"`
	Vars   DeclarationList `json:",omitempty"`
	Funcs  DeclarationList `json:",omitempty"`
}

// DeclarationList merges all the declarations together.
func (d *Definition) DeclarationList() DeclarationList {
	allDecls := DeclarationList{}
	allDecls.Append(d.Consts...)
	allDecls.Append(d.Vars...)
	allDecls.Append(d.Types...)
	allDecls.Append(d.Funcs...)
	return allDecls
}

// Sort will sort the inner types so they have a stable order.
func (d *Definition) Sort() {
	d.Types.Sort()
	d.Vars.Sort()
	d.Consts.Sort()
	d.Funcs.Sort()
}

func (d *Definition) Order() DeclarationList {
	count := len(d.Types) + len(d.Funcs) + len(d.Vars) + len(d.Consts)
	result := make([]*Declaration, 0, count)

	result = append(result, d.Types...)
	result = append(result, d.Funcs...)
	result = append(result, d.Vars...)
	result = append(result, d.Consts...)
	return DeclarationList(result)
}

func (d *Definition) ClearSource() {
	d.Types.ClearSource()
	d.Vars.ClearSource()
	d.Consts.ClearSource()
	d.Funcs.ClearSource()
}

func (d *Definition) ClearTestFiles() {
	for filename, _ := range d.Imports {
		if strings.HasSuffix(filename, "_test.go") {
			delete(d.Imports, filename)
		}
	}
	d.Types.ClearTestFiles()
	d.Vars.ClearTestFiles()
	d.Consts.ClearTestFiles()
	d.Funcs.ClearTestFiles()
}

func (d *Definition) ClearNonTestFiles() {
	for filename, _ := range d.Imports {
		if !strings.HasSuffix(filename, "_test.go") {
			delete(d.Imports, filename)
		}
	}
	d.Types.ClearNonTestFiles()
	d.Vars.ClearNonTestFiles()
	d.Consts.ClearNonTestFiles()
	d.Funcs.ClearNonTestFiles()
}

func (d *Definition) Fill() {
	for _, decl := range d.Order() {
		decl.Imports = d.getImports(decl)
	}
}

func (d *Definition) Merge(in *Definition) {
	d.TestPackage = d.TestPackage || in.TestPackage

	for k, v := range in.Imports {
		d.Imports.Add(k, v...)
	}

	d.Types.AppendUnique(in.Types...)
	d.Funcs.AppendUnique(in.Funcs...)
	d.Vars.AppendUnique(in.Vars...)
	d.Consts.AppendUnique(in.Consts...)

	// this line causes Sort to be omitted from the
	// definitions :/ ... solved by adding the sort
	// in the AppendUnique above, but the Sort symbol
	// should not be omitted from Definition.

	// d.Sort()
}

func (d *Definition) getImports(decl *Declaration) []string {
	return d.Imports.Get(decl.File)
}
