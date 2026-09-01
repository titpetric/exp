package restore

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/titpetric/exp/cmd/go-fsck/internal"
	"github.com/titpetric/exp/cmd/go-fsck/internal/files"
	"github.com/titpetric/tools/splint/model"
)

// restore writes every package of the document back out as source.
//
// A package is written at the path it was extracted from, under the output
// path, so a document of one package restores into one directory and a
// document of a tree restores as the tree. Nothing is rewritten on the way:
// the package clause and the imports are what the source said, and a package
// reaching another package of the same tree reaches it by the path it was
// published under.
func restore(cfg *options) error {
	defs, err := internal.ReadDocument(cfg.inputFile)
	if err != nil {
		return err
	}
	if len(defs) == 0 {
		return fmt.Errorf("%s holds no packages", cfg.inputFile)
	}

	kinds, err := cfg.kinds()
	if err != nil {
		return err
	}

	// A directory holds the package and the external test package beside it,
	// and the two are written together so a filename one takes is a filename
	// the other knows about.
	for _, dir := range directories(defs) {
		if err := restoreDirectory(cfg, kinds, dir); err != nil {
			return err
		}
	}

	return nil
}

// directory is every definition read from one package directory.
type directory struct {
	path string
	defs []*model.Definition
}

// directories groups the document by the directory each definition was read
// from, in the order the document lists them.
func directories(defs model.DefinitionList) []*directory {
	var out []*directory
	at := map[string]*directory{}

	for _, def := range defs {
		dir, known := at[def.Package.Path]
		if !known {
			dir = &directory{path: def.Package.Path}
			at[def.Package.Path] = dir
			out = append(out, dir)
		}
		dir.defs = append(dir.defs, def)
	}

	return out
}

// restoreDirectory writes the files of one package directory.
func restoreDirectory(cfg *options, kinds map[model.DeclarationKind]bool, dir *directory) error {
	target := filepath.Join(cfg.outputPath, relative(dir.path))
	if err := os.MkdirAll(target, 0o755); err != nil {
		return err
	}

	// The name of a file is taken by the package that claimed it first. The
	// external test package is a package of its own in the same directory, so
	// what it writes is named apart from what the package under test writes.
	taken := map[string]string{}

	// The layout is of the directory rather than of one package in it: a test
	// is placed against the symbol it covers, and the symbol is declared in
	// the package the tests are of.
	shared := newLayout(main(dir), cfg.split, everything(cfg, kinds, dir))

	for _, def := range dir.defs {
		decls := wanted(cfg, kinds, def)
		if len(decls) == 0 {
			continue
		}

		pkg := clause(def)
		out := shared.forPackage(pkg)
		buckets := map[string]model.DeclarationList{}
		order := []string{}

		for _, decl := range decls {
			name := claim(taken, out.File(decl), pkg)
			if _, known := buckets[name]; !known {
				order = append(order, name)
			}
			buckets[name] = append(buckets[name], decl)
		}

		sort.Strings(order)
		for _, name := range order {
			if err := writeBucket(cfg, def, pkg, target, name, buckets[name]); err != nil {
				return err
			}
		}
	}

	return writeDoc(cfg, target, dir)
}

// main is the package clause of the directory, which is the one the tests in
// it are tests of.
func main(dir *directory) string {
	for _, def := range dir.defs {
		if !def.Package.TestPackage {
			return def.Package.Package
		}
	}
	if len(dir.defs) > 0 {
		return clause(dir.defs[0])
	}
	return ""
}

// everything are the declarations of every package of the directory, which is
// what a test is placed against.
func everything(cfg *options, kinds map[model.DeclarationKind]bool, dir *directory) model.DeclarationList {
	var out model.DeclarationList
	for _, def := range dir.defs {
		out = append(out, wanted(cfg, kinds, def)...)
	}
	return out
}

// wanted are the declarations of a definition the run asked for, in the order
// the document lists them.
func wanted(cfg *options, kinds map[model.DeclarationKind]bool, def *model.Definition) model.DeclarationList {
	var out model.DeclarationList

	for _, decl := range def.DeclarationList() {
		switch {
		case kinds != nil && !kinds[decl.Kind]:
		case cfg.noTests && decl.IsTestScope():
		case cfg.removeUnexported && !decl.IsExported():
		default:
			out = append(out, decl)
		}
	}

	return out
}

// writeBucket writes one file: the declarations it holds, and the imports they
// reach.
func writeBucket(cfg *options, def *model.Definition, pkg, target, name string, decls model.DeclarationList) error {
	decls.Sort()

	sources := make([]string, 0, len(decls))
	for _, decl := range decls {
		if decl.Source == "" {
			return fmt.Errorf("%s has no source for %s: extract it with --include-sources",
				cfg.inputFile, decl.Symbol())
		}
		sources = append(sources, decl.Source)
	}

	file := &files.File{
		Filename: filepath.Join(target, name),
		Package:  pkg,
		Types:    sources,
	}

	imports, err := fileImports(def, decls, file.Body())
	if err != nil {
		return fmt.Errorf("%s: %w", name, err)
	}
	file.Imports = imports

	if cfg.verbose {
		fmt.Printf("%s: %d symbols, %d imports\n", file.Filename, len(decls), len(imports))
	}

	return file.Flush()
}

// clause is the package clause a definition was written under.
//
// The model names both test scopes of a package "x_test", because they are two
// scopes and the listing tells them apart that way. The source does not: the
// tests inside a package are written "package x" and the tests beside it
// "package x_test", and what separates them is that the second imports the
// first. That is what is read here, because a file written under the wrong
// clause is a file that does not compile.
func clause(def *model.Definition) string {
	pkg := def.Package.Package
	if !def.Package.TestPackage || !strings.HasSuffix(pkg, "_test") {
		return pkg
	}

	under := strings.TrimSuffix(def.Package.ImportPath, "_test")
	for _, literals := range def.Imports {
		for _, literal := range literals {
			if strings.Trim(literal[strings.LastIndex(literal, " ")+1:], `"`) == under {
				return pkg
			}
		}
	}

	return strings.TrimSuffix(pkg, "_test")
}

// writeDoc writes the package comment, which belongs to the package rather
// than to any symbol of it.
func writeDoc(cfg *options, target string, dir *directory) error {
	for _, def := range dir.defs {
		if def.Doc == "" || def.Package.TestPackage {
			continue
		}

		file := &files.File{
			Filename: filepath.Join(target, "doc.go"),
			Package:  def.Package.Package,
			Doc:      def.Doc,
		}
		if cfg.verbose {
			fmt.Println(file.Filename + ": the package comment")
		}
		return file.Flush()
	}

	return nil
}

// claim returns the name a package writes a file under, given what the
// packages before it in the same directory took.
//
// Two packages of one directory can want one name: the tests inside a package
// and the tests beside it both collect what they do not name in a file named
// for the package. The second one to ask is told apart.
func claim(taken map[string]string, name, pkg string) string {
	owner, known := taken[name]
	if !known {
		taken[name] = pkg
		return name
	}
	if owner == pkg {
		return name
	}

	apart := strings.TrimSuffix(name, "_test.go") + "_ext_test.go"
	taken[apart] = pkg
	return apart
}

// relative is a package path as a directory under the output path, with the
// dot the document writes it with taken off.
func relative(path string) string {
	path = strings.TrimPrefix(path, "./")
	if path == "." || path == "" {
		return ""
	}
	return path
}
