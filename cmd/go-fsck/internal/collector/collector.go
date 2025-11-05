package collector

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/printer"
	"go/token"
	"log"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"

	. "github.com/titpetric/exp/cmd/go-fsck/internal/ast"
	"github.com/titpetric/exp/cmd/go-fsck/model"
)

type (
	Definition  = model.Definition
	Declaration = model.Declaration
	Package     = model.Package
)

type collector struct {
	fset *token.FileSet

	definition map[string]*Definition
	seen       map[string]bool
}

func NewCollector(fset *token.FileSet) *collector {
	return &collector{
		fset:       fset,
		definition: make(map[string]*Definition),
		seen:       make(map[string]bool),
	}
}

func (v *collector) Clean(verbose bool) []*Definition {
	for _, def := range v.definition {
		importMap, _ := def.Imports.Map()

		for _, fv := range def.Funcs {
			for k, v := range fv.References {
				if _, ok := importMap[k]; !ok {
					for _, p := range v {
						def.References.Add(fmt.Sprintf("%s.%s", k, p), fv.Name)
					}
					delete(fv.References, k)
				}
			}
		}
	}

	results := make([]*Definition, 0, len(v.definition))
	pkgNames := make([]string, 0, len(v.definition))
	for _, pkg := range v.definition {
		pkg.Sort()
		pkgNames = append(pkgNames, pkg.Package.Path)
	}
	sort.Strings(pkgNames)

	for _, pkg := range v.definition {
		for _, name := range pkgNames {
			if pkg.Package.Path == name {
				results = append(results, pkg)
			}
		}
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].Package.Path < results[j].Package.Path
	})

	return results
}

func (v *collector) setSeen(key string) {
	// fmt.Printf("seen: %s\n", key)
	v.seen[key] = true
}

func (v *collector) isSeen(key string) bool {
	_, ok := v.seen[key]
	return ok
}

func (v *collector) collectImports(filename string, decl *ast.GenDecl, def *Definition) {
	for _, spec := range decl.Specs {
		imported, ok := spec.(*ast.ImportSpec)
		if !ok {
			continue
		}

		importLiteral := imported.Path.Value
		importClean := strings.Trim(importLiteral, `*`)
		if imported.Name != nil {
			alias := imported.Name.Name
			base := path.Base(importClean)
			switch alias {
			case base:
				fmt.Fprintf(os.Stderr, "WARN: removing %s alias for %s)\n", alias, importClean)
			case "_":
				// no warning
			default:
				// fmt.Printf("WARN: package %s is aliased to %s\n", importLiteral, alias)
				importLiteral = alias + " " + importLiteral
			}
		}

		def.Imports.Add(filepath.Base(filename), importLiteral)
	}
}

func collectFuncReferences(funcDecl *ast.FuncDecl) map[string][]string {
	imports := make(map[string][]string)

	if funcDecl == nil || funcDecl.Body == nil {
		return imports
	}

	// Traverse the function body and look for package identifiers.
	ast.Inspect(funcDecl.Body, func(node ast.Node) bool {
		switch n := node.(type) {
		case *ast.SelectorExpr:
			// If it's a SelectorExpr, get the leftmost identifier which is the package name.
			if ident, ok := n.X.(*ast.Ident); ok {
				pkgName := ident.Name

				if ident.Obj != nil {
					if ident.Obj.Kind != ast.Pkg {
						// pkgName is not a package
						return true
					}
				}

				selName := n.Sel.Name
				if pkgName != "internal" && ast.IsExported(selName) {
					imports[pkgName] = appendIfNotExists(imports[pkgName], selName)
				}
			}
		case *ast.Ident:
			// If it's an identifier, it might be a package name.
			if obj := n.Obj; obj != nil && obj.Kind == ast.Pkg {
				pkgName := n.Name
				imports[pkgName] = nil // No specific symbol, just mark the package as imported.
			}
		}

		return true
	})

	return imports
}

func (v *collector) Visit(node ast.Node, push bool, stack []ast.Node) bool {
	file, ok := (stack[0]).(*ast.File)
	if !ok {
		return true
	}
	filename := v.fset.Position(file.Pos()).Filename

	packageName := file.Name.Name

	pkg, ok := v.definition[packageName]
	if !ok {
		pkg = &Definition{}
		pkg.Package.Path = filepath.Dir(filename)
		pkg.Package.Package = packageName

		v.definition[packageName] = pkg
	}

	if file.Doc != nil {
		pkg.Doc = strings.TrimSpace(v.getSource(file, file.Doc))
	}

	switch node := node.(type) {
	case *ast.GenDecl:
		if node.Tok == token.IMPORT {
			v.collectImports(filename, node, pkg)
			return true
		}

		// If there's a function declaration in the stack,
		// the var/const/struct is internal to a function.
		for _, k := range stack {
			_, ok := k.(*ast.FuncDecl)
			if ok {
				return true
			}
		}

		names := v.Names(node)

		for _, name := range names {
			if v.isSeen("decl:" + packageName + "." + name) {
				return true
			}
		}

		def := &Declaration{
			Names:         names,
			File:          filepath.Base(filename),
			SelfContained: IsSelfContainedType(node),
			Source:        v.getSource(file, node),
		}
		if len(def.Names) == 1 {
			def.Name = def.Names[0]
			def.Names = nil

			def.Doc = strings.TrimSpace(v.getSource(file, node.Doc))
		}

		for _, name := range names {
			v.setSeen("decl:" + packageName + "." + name)
		}

		switch node.Tok {
		case token.CONST:
			def.Kind = model.ConstKind
			pkg.Consts.Append(def)
		case token.VAR:
			def.Kind = model.VarKind
			pkg.Vars.Append(def)
		case token.TYPE:
			def.Kind = model.TypeKind

			v.collectStructFields(def, file, node)

			pkg.Types.Append(def)
		}

	case *ast.FuncDecl:
		// func name
		name := node.Name.Name
		if node.Recv != nil {
			// func receiver.name
			name = v.symbolType(file, node.Recv.List[0].Type) + "." + name
		}

		key := name
		if packageName != "" {
			key = packageName + "." + name
		}
		if v.isSeen("func:" + key) {
			return true
		}
		v.setSeen("func:" + key)

		def := v.collectFuncDeclaration(file, node, filename, stack)
		if def != nil {
			pkg.Funcs.Append(def)
		}
	}

	return true
}

func (v *collector) Names(decl *ast.GenDecl) []string {
	names := make([]string, 0, len(decl.Specs))
	for _, spec := range decl.Specs {
		if val, ok := spec.(*ast.ValueSpec); ok {
			names = append(names, v.identNames(val.Names)...)
			continue
		}

		if val, ok := spec.(*ast.TypeSpec); ok {
			names = append(names, val.Name.Name)
			continue
		}

		v.error("warning getting names: unhandled %T", spec)
	}
	if len(names) == 0 {
		return nil
	}
	return names
}

func (v *collector) error(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, format+"\n", args...)
}

func (v *collector) identNames(decl []*ast.Ident) []string {
	if len(decl) == 0 {
		return nil
	}

	result := make([]string, 0, len(decl))
	for _, t := range decl {
		result = append(result, t.Name)
	}
	return result
}

func (v *collector) collectFuncDeclaration(file *ast.File, decl *ast.FuncDecl, filename string, stack []ast.Node) *Declaration {
	args, returns := v.functionBindings(file, decl)

	source := v.getSource(file, decl)

	declaration := &Declaration{
		Doc:        strings.TrimSpace(v.getSource(file, decl.Doc)),
		Kind:       model.FuncKind,
		File:       filepath.Base(filename),
		Name:       decl.Name.Name,
		Arguments:  args,
		Returns:    returns,
		Signature:  v.functionDef(decl),
		References: collectFuncReferences(decl),
		Source:     source,
		Complexity: complexity(decl, source),
	}

	if decl.Recv != nil {
		declaration.Receiver = v.symbolType(file, decl.Recv.List[0].Type)
	}

	return declaration
}

func (p *collector) getSource(file *ast.File, node any) string {
	if commentGroup, ok := node.(*ast.CommentGroup); ok {
		return commentGroup.Text()
	}

	var buf strings.Builder
	err := PrintSource(&buf, p.fset, CommentedNode(file, node))
	if err != nil {
		fmt.Printf("Error printing source: %v\n", err)
		return ""
	}
	return buf.String()
}

func (p *collector) symbolType(file *ast.File, expr ast.Expr) string {
	switch t := expr.(type) {
	case *ast.Ident:
		return t.Name
	case *ast.StarExpr:
		return "*" + p.symbolType(file, t.X)
	case *ast.ArrayType:
		return "[]" + p.symbolType(file, t.Elt)
	case *ast.Ellipsis:
		return "..." + p.symbolType(file, t.Elt)
	case *ast.SelectorExpr:
		return p.symbolType(file, t.X) + "." + p.symbolType(file, t.Sel)
	case *ast.MapType:
		var (
			k = p.symbolType(file, t.Key)
			v = p.symbolType(file, t.Value)
		)
		return fmt.Sprintf("map[%s]%s", k, v)
	case *ast.InterfaceType:
		return "any"
	}
	return p.getSource(file, expr)
}

func (p *collector) functionBindings(file *ast.File, decl *ast.FuncDecl) (args []string, returns []string) {
	// Traverse arguments
	for _, field := range decl.Type.Params.List {
		argType := p.symbolType(file, field.Type)
		args = appendIfNotExists(args, argType)
	}

	// Traverse return values
	if decl.Type.Results != nil {
		for _, field := range decl.Type.Results.List {
			returnType := p.symbolType(file, field.Type)
			returns = appendIfNotExists(returns, returnType)
		}
	}
	return
}

func (p *collector) functionDef(fun *ast.FuncDecl) string {
	var fset = p.fset
	name := fun.Name.Name
	params := make([]string, 0)
	for _, p := range fun.Type.Params.List {
		var typeNameBuf bytes.Buffer
		err := printer.Fprint(&typeNameBuf, fset, p.Type)
		if err != nil {
			log.Fatalf("failed printing %s", err)
		}

		names := make([]string, 0, len(p.Names))
		for _, name := range p.Names {
			names = append(names, name.Name)
		}

		if len(names) > 0 {
			params = append(params, fmt.Sprintf("%s %s", strings.Join(names, ","), typeNameBuf.String()))
		} else {
			// unnamed parameter: parts of interfaces, unused parameter, like `_` but omitted name.
			params = append(params, typeNameBuf.String())
		}
	}
	returns := make([]string, 0)
	if fun.Type.Results != nil {
		for _, r := range fun.Type.Results.List {
			var typeNameBuf bytes.Buffer
			err := printer.Fprint(&typeNameBuf, fset, r.Type)
			if err != nil {
				log.Fatalf("failed printing %s", err)
			}

			returns = append(returns, typeNameBuf.String())
		}
	}
	returnString := ""
	if len(returns) == 1 {
		returnString = returns[0]
	} else if len(returns) > 1 {
		returnString = fmt.Sprintf("(%s)", strings.Join(returns, ", "))
	}

	paramsString := strings.Join(params, ", ")
	if returnString != "" {
		return fmt.Sprintf("%s (%s) %s", name, paramsString, returnString)
	}
	return fmt.Sprintf("%s (%s)", name, paramsString)
}

func appendIfNotExists(slice []string, element string) []string {
	for _, s := range slice {
		if s == element {
			return slice
		}
	}
	return append(slice, element)
}
