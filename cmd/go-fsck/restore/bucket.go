package restore

import (
	"strings"

	"github.com/titpetric/tools/splint/model"
)

// prefixes are what a test is named with, and come off before the name of what
// it tests is read: TestTrace tests Trace, and belongs beside it.
var prefixes = []string{"Test", "Benchmark", "Example", "Fuzz"}

// layout is which file each declaration of one package goes in.
//
// It is built in two passes because the answer depends on the types: a method
// goes where its receiver is, a constructor where what it returns is, and a
// const typed as one of them beside the type it is typed as. The types are
// placed first, and everything else is placed against them.
type layout struct {
	// pkg is the package clause, which names the file the symbols that get no
	// file of their own are collected in.
	pkg string

	// split puts every symbol in the file named for it, unexported ones
	// included.
	split bool

	// types is the file each type of the package was placed in, and symbols
	// the file every symbol of it was placed in, both keyed by the name the
	// symbol is declared under.
	//
	// The types are what a method, a constructor and a typed value are placed
	// against, and the symbols are what a test is placed against.
	types   map[string]string
	symbols map[string]string
}

// newLayout places the types of a package, which is what everything else is
// placed against.
func newLayout(pkg string, split bool, decls model.DeclarationList) *layout {
	out := &layout{
		pkg:     pkg,
		split:   split,
		types:   map[string]string{},
		symbols: map[string]string{},
	}

	for _, decl := range decls {
		if decl.Kind != model.TypeKind || decl.IsTestScope() {
			continue
		}
		for _, name := range decl.GetNames() {
			if name != "" {
				out.types[name] = out.named(name)
			}
		}
	}

	// Everything else is placed against the types, and a test is placed
	// against everything else.
	for _, decl := range decls {
		if decl.IsTestScope() {
			continue
		}
		for _, name := range decl.GetNames() {
			if name != "" {
				out.symbols[name] = out.place(decl, name)
			}
		}
	}

	return out
}

// forPackage is the layout as one package of the directory sees it: the same
// symbols, and the file the ones with no file of their own are collected in.
func (l *layout) forPackage(pkg string) *layout {
	out := *l
	out.pkg = pkg
	return &out
}

// named is the file a symbol of this name is written in, which is the file
// named for it when it has one of its own and the package file otherwise.
func (l *layout) named(name string) string {
	if name == "" {
		return l.pkg + ".go"
	}
	if !l.split && !isExported(name) {
		return l.pkg + ".go"
	}
	return toFilename(name)
}

// File is where one declaration goes.
//
// A test declaration goes where the symbol it names goes, with the test suffix
// on it: what a test covers is the file it belongs beside, and a test the
// naming does not resolve joins the package file.
func (l *layout) File(decl *model.Declaration) string {
	name := decl.Name
	if name == "" && len(decl.Names) > 0 {
		name = decl.Names[0]
	}

	if decl.IsTestScope() {
		return testFile(l.tested(name))
	}
	return l.place(decl, name)
}

// place is the file a declaration goes in, reading the name it was asked to
// place it under.
func (l *layout) place(decl *model.Declaration, name string) string {
	switch decl.Kind {
	case model.FuncKind:
		return l.function(decl, name)
	case model.TypeKind:
		return l.named(name)
	case model.ConstKind, model.VarKind:
		return l.value(decl, name)
	}
	return l.pkg + ".go"
}

// function is the file a func goes in: beside its receiver where it has one,
// beside what it constructs where it is a constructor, and in the file named
// for it otherwise.
func (l *layout) function(decl *model.Declaration, name string) string {
	if receiver := model.TypeRef(decl.Receiver); receiver != "" {
		if file, known := l.types[receiver]; known {
			return file
		}
		return l.named(receiver)
	}

	// A constructor is named for what it returns, so New comes off the front
	// before the name is looked up: NewTracer belongs with Tracer.
	if constructed := strings.TrimPrefix(name, "New"); constructed != name {
		if file, known := l.types[constructed]; known {
			return file
		}
	}

	return l.named(name)
}

// value is the file a const or a var goes in: beside the type it is declared
// as where the package declares one, and in the file the values of the package
// are collected in otherwise.
//
// A method and a constructor stay with their type under split, because what
// they are is part of the type. A value is not, so it gets a file of its own.
func (l *layout) value(decl *model.Declaration, name string) string {
	// Under split every symbol is in the file named for it, and there is no
	// file the rest are collected in.
	if l.split {
		return l.named(name)
	}

	if typ := model.TypeRef(decl.Type); typ != "" {
		if file, known := l.types[typ]; known {
			return file
		}
	}

	if isExported(name) {
		if decl.Kind == model.ConstKind {
			return "const.go"
		}
		return "vars.go"
	}

	return l.pkg + ".go"
}

// tested is the file the symbol a test covers was placed in.
//
// The name is read back a word at a time, so TestTracerUsesConfiguredStorage
// finds Tracer the way splint's coverage linter does. A name that resolves to
// nothing the package declares is a test of the package rather than of a
// symbol, and joins the file named for it.
func (l *layout) tested(name string) string {
	name = testedName(name)

	for words := splitCamel(name); len(words) > 0; words = words[:len(words)-1] {
		if file, known := l.symbols[strings.Join(words, "")]; known {
			return file
		}
	}

	return l.pkg + ".go"
}

// splitCamel breaks an identifier on its capitals, so TracerUsesStorage comes
// back as Tracer, Uses and Storage.
func splitCamel(name string) []string {
	var (
		words []string
		start int
	)

	for i, r := range name {
		if i > 0 && r >= 'A' && r <= 'Z' {
			words = append(words, name[start:i])
			start = i
		}
	}
	if start < len(name) {
		words = append(words, name[start:])
	}

	return words
}

// testedName is the name a test names, which is what it is called less the
// prefix the toolchain reads it by and whatever follows the underscore:
// TestTracer_Live covers Tracer.
func testedName(name string) string {
	for _, prefix := range prefixes {
		if trimmed := strings.TrimPrefix(name, prefix); trimmed != name {
			name = trimmed
			break
		}
	}

	if under := strings.Index(name, "_"); under > 0 {
		name = name[:under]
	}
	return name
}

// testFile is the test file beside a file, and is the file itself where that
// is already a test file: the tests beside a package are a package named for
// the tests, and the file they collect in is named for the package.
func testFile(name string) string {
	if strings.HasSuffix(name, "_test.go") {
		return name
	}
	return strings.TrimSuffix(name, ".go") + "_test.go"
}

// isExported reports a name a reader outside the package can reach.
func isExported(name string) bool {
	return name != "" && name[0] >= 'A' && name[0] <= 'Z'
}
