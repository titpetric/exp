package extract

import (
	"bytes"
	"errors"
	"fmt"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"log"
	"os"
	"path"
	"sort"
	"strings"

	"golang.org/x/exp/slices"

	"github.com/titpetric/exp/cmd/schema-gen/model"
)

func NewExtractOptions(cfg *options) *model.ExtractOptions {
	return &model.ExtractOptions{
		IncludeFunctions:  cfg.includeFunctions,
		IncludeTests:      cfg.includeTests,
		IncludeUnexported: cfg.includeUnexported,
		IgnoreFiles:       cfg.ignoreFiles,
		IncludeInternal:   cfg.includeInternal,
	}
}

// Extract package structs
func Extract(filepath string, options *model.ExtractOptions) ([]*model.PackageInfo, error) {
	var (
		ignoreFiles = options.IgnoreFiles
	)

	ignoreList := make(map[string]bool)
	for _, file := range ignoreFiles {
		ignoreList[file] = true
	}

	// filter skips explicitly ignored files, and tests files
	filter := func(fInfo os.FileInfo) bool {
		if ignored := ignoreList[fInfo.Name()]; ignored {
			return false
		}

		if strings.HasSuffix(fInfo.Name(), "_test.go") {
			return options.IncludeTests
		}

		return true
	}

	fileSet := token.NewFileSet()
	pkgs, err := parser.ParseDir(fileSet, path.Dir(filepath), filter, parser.ParseComments)
	if err != nil {
		return nil, err
	}

	result := make([]*model.PackageInfo, 0, len(pkgs))
	for pkgName, pkg := range pkgs {
		p := NewParser(fileSet, pkg)
		pkgInfo, err := p.GetDeclarations(options)
		pkgInfo.Name = pkgName
		if err != nil {
			return nil, err
		}
		result = append(result, pkgInfo)
	}
	return result, nil
}

type objParser struct {
	pkg     *ast.Package
	fileset *token.FileSet

	visited map[string]bool // avoid re-visiting struct type
}

func NewParser(fileset *token.FileSet, pkg *ast.Package) *objParser {
	p := &objParser{
		pkg:     pkg,
		fileset: fileset,
		visited: map[string]bool{},
	}
	return p
}

func (p *objParser) getSourceCode(node ast.Node) string {
	var buf strings.Builder
	err := printer.Fprint(&buf, p.fileset, node)
	if err != nil {
		return ""
	}
	return buf.String()
}

func (p *objParser) functionDef(fun *ast.FuncDecl) string {
	var fset = p.fileset
	name := fun.Name.Name
	params := make([]string, 0)
	for _, p := range fun.Type.Params.List {
		var typeNameBuf bytes.Buffer
		err := printer.Fprint(&typeNameBuf, fset, p.Type)
		if err != nil {
			log.Fatalf("failed printing %s", err)
		}
		names := make([]string, 0)
		for _, name := range p.Names {
			names = append(names, name.Name)
		}
		params = append(params, fmt.Sprintf("%s %s", strings.Join(names, ","), typeNameBuf.String()))
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
		return fmt.Sprintf("%s (%s) %v", name, paramsString, returnString)
	}
	return fmt.Sprintf("%s (%s)", name, paramsString)
}

// GetDeclaration returns a filled out PackageInfo{} and an error if any.
func (p *objParser) GetDeclarations(options *model.ExtractOptions) (*model.PackageInfo, error) {
	var err error
	result := &model.PackageInfo{
		Declarations: model.DeclarationList{},
		Imports:      []string{},
	}

	var funcs []*model.FuncInfo
	var globalFuncs []*model.FuncInfo

	for _, fileObj := range p.pkg.Files {
		// https://pkg.go.dev/go/ast#File

		if options.IncludeFunctions {
			ast.Inspect(fileObj, func(n ast.Node) (res bool) {
				res = true
				if fun, ok := n.(*ast.FuncDecl); ok {
					if !fun.Name.IsExported() && !options.IncludeUnexported {
						return
					}
					if fun.Recv == nil || len(fun.Recv.List) == 0 {
						name := getTypeDeclaration(fun.Name)
						funcinfo := &model.FuncInfo{
							Name:      name,
							Doc:       TrimSpace(fun.Doc),
							Path:      name,
							Signature: p.functionDef(fun),
							Source:    p.getSourceCode(n),
						}
						globalFuncs = append(globalFuncs, funcinfo)
						return
					}
					if r, ok := fun.Recv.List[0].Type.(*ast.StarExpr); ok {
						if len(fun.Recv.List[0].Names) == 0 {
							return
						}
						recvType := getTypeDeclaration(fun.Recv.List[0].Names[0]) + " " + getTypeDeclarationsForPointerType(r)
						signature := p.functionDef(fun)
						goPath := r.X.(*ast.Ident).Name
						funcinfo := &model.FuncInfo{
							Name:      getTypeDeclaration(fun.Name),
							Doc:       TrimSpace(fun.Doc),
							Type:      recvType,
							Path:      goPath,
							Signature: signature,
							Source:    p.getSourceCode(n),
						}
						funcs = append(funcs, funcinfo)
					}

				}

				return
			})
		}

		// Collect imports
		for _, imported := range fileObj.Imports {
			importLiteral := imported.Path.Value
			if imported.Name != nil {
				alias := getTypeDeclaration(imported.Name)
				//fmt.Fprintf(os.Stderr, "WARN: package %s is aliased to %s\n", importLiteral, alias)
				importLiteral = alias + " " + importLiteral
			}

			if !strings.Contains(importLiteral, "/internal") || options.IncludeInternal {
				if slices.Contains(result.Imports, importLiteral) {
					continue
				}
				result.Imports = append(result.Imports, importLiteral)
			}
		}

		// Collect declarations
		for _, obj := range fileObj.Decls {
			genDecl, ok := obj.(*ast.GenDecl)
			if !ok {
				continue
			}

			info := &model.DeclarationInfo{
				Doc:     TrimSpace(genDecl.Doc),
				FileDoc: TrimSpace(fileObj.Doc),
				Types:   model.TypeList{},
			}

			for _, spec := range genDecl.Specs {
				switch obj := spec.(type) {
				case *ast.ValueSpec, *ast.ImportSpec:
				case *ast.TypeSpec:
					typeInfo, err := NewTypeSpecInfo(obj)
					if err != nil {
						isUnexported := errors.Is(err, model.ErrUnexported)
						if isUnexported && !options.IncludeUnexported {
							continue
						}
					}

					p.parseStruct(typeInfo.Name, typeInfo.Name, typeInfo, options)

					info.Types.Append(typeInfo)
				default:
					fmt.Fprintf(os.Stderr, "INFO: Unhandled AST %T\n", spec)
				}
			}

			if info.Valid() {
				result.Declarations.Append(info)
			}

			var currentType string
			var currentValue any = 0

			for _, spec := range genDecl.Specs {
				switch obj := spec.(type) {
				case *ast.TypeSpec, *ast.ImportSpec:
				case *ast.ValueSpec:
					if ident, ok := obj.Type.(*ast.Ident); ok {
						currentType = ident.Name
					}

					if currentType == "" {
						break
					}

					typeMap := result.Declarations.TypeMap()
					typeInfo, ok := typeMap[currentType]
					if !ok || typeInfo == nil {
						// fmt.Fprintf(os.Stderr, "INFO: Unknown type in TypeMap: looking for %q\n", currentType)
						continue
					}

					for i, name := range obj.Names {
						enumValue := &model.EnumInfo{
							Name:  name.Name,
							Doc:   TrimSpace(obj.Doc),
							Value: currentValue,
						}

						if len(obj.Values) > i {
							if basicLit, ok := obj.Values[i].(*ast.BasicLit); ok {
								switch basicLit.Kind {
								case token.INT:
									var val int
									fmt.Sscanf(basicLit.Value, "%d", &val)
									currentValue = val
									enumValue.Value = val
								case token.STRING:
									val := strings.Trim(basicLit.Value, "\"")
									currentValue = val
									enumValue.Value = val
								}
							}
						}

						// Increment only if currentValue is int.
						// This doesn't handle `1 << enum` type declarations.
						if v, ok := currentValue.(int); ok {
							currentValue = v + 1
						}

						typeInfo.Enums = append(typeInfo.Enums, enumValue)
					}
				default:
					// Type already logged in first loop.
				}
			}

			sort.Stable(info.Types)

		}
	}

	if options.IncludeFunctions {
		for _, funcInfo := range funcs {
			for _, decl := range result.Declarations {
				for _, typeDecl := range decl.Types {
					if typeDecl.Name == funcInfo.Path {
						typeDecl.Functions = append(typeDecl.Functions, funcInfo)
					}
				}
			}
		}

		result.Functions = globalFuncs
	}

	sort.Stable(result.Declarations)
	slices.Sort(result.Imports)

	return result, err
}

// NewTypeSpecInfo allocates a TypeInfo from a TypeSpec node.
// The function returns the struct info and ErrUnexported if
// the struct is not exported. This error is handled outside
// to skip documenting unexported types.
func NewTypeSpecInfo(from *ast.TypeSpec) (*model.TypeInfo, error) {
	info := &model.TypeInfo{
		Name:    getTypeDeclaration(from.Name),
		Doc:     TrimSpace(from.Doc),
		Comment: TrimSpace(from.Comment),
		Enums:   []*model.EnumInfo{},
	}

	structObj, ok := from.Type.(*ast.StructType)
	if ok {
		info.StructObj = structObj
	} else {
		info.Type = TypeName(from)
	}

	if !ast.IsExported(info.Name) {
		return info, model.ErrUnexported
	}
	return info, nil
}

func (p *objParser) parseStruct(goPath, name string, structInfo *model.TypeInfo, options *model.ExtractOptions) {
	if structInfo.StructObj == nil {
		return
	}
	if visited, _ := p.visited[name]; visited {
		return
	}
	p.visited[name] = true

	for _, field := range structInfo.StructObj.Fields.List {
		//pos := p.fileset.Position(field.Pos())
		//filePos := path.Base(pos.String())

		var goName string
		if len(field.Names) > 0 {
			goName = field.Names[0].Name

		}

		tagValue := ""
		if field.Tag != nil {
			tagValue = string(field.Tag.Value)
			tagValue = strings.Trim(tagValue, "`")
		}

		jsonName := jsonTag(tagValue)
		if jsonName == "" {
			// fields without json tag encode to field name
			jsonName = goName
		}
		if jsonName == "-" {
			// fields with json `-` don't get encoded
			jsonName = ""
		}

		fieldPath := goName
		if goPath != "" {
			fieldPath = goPath
			if goName != "" {
				fieldPath += "." + goName
			}
		}

		fieldInfo := &model.FieldInfo{
			Doc:     TrimSpace(field.Doc),
			Comment: TrimSpace(field.Comment),

			Name: goName,
			Path: fieldPath,
			Type: getTypeDeclarationsForExpr(field.Type),
			Tag:  tagValue,

			JSONName: jsonName,
		}

		isExported := ast.IsExported(fieldInfo.Name)
		if !isExported && !options.IncludeUnexported {
			continue
		}

		structInfo.Fields = append(structInfo.Fields, fieldInfo)
	}
	return
}
