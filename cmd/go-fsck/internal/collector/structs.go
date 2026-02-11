package collector

import (
	"go/ast"
	"strings"

	. "github.com/titpetric/exp/cmd/go-fsck/internal/ast"
	"github.com/titpetric/exp/cmd/go-fsck/model"
)

func (p *collector) collectStructFields(out *model.Declaration, file *ast.File, decl *ast.GenDecl) {
	for _, spec := range decl.Specs {
		switch obj := spec.(type) {
		case *ast.TypeSpec:
			switch val := obj.Type.(type) {
			case *ast.StructType:
				p.parseStruct(out, file, obj, val)
			case *ast.InterfaceType:
				out.Type = "interface"
				for _, field := range val.Methods.List {
					if len(field.Names) == 0 {
						out.Fields = append(out.Fields, &model.Field{
							Type: "interface",
						})
						continue
					}
					for _, name := range field.Names {
						//fmt.Println(name, p.functionType(name.Name, field.Type.(*ast.FuncType)))
						out.Fields = append(out.Fields, &model.Field{
							Name: name.Name,
							Type: p.functionType(name.Name, field.Type.(*ast.FuncType)),
						})
					}
				}
			case *ast.Ident:
				// type aliases, e.g. type X string
			default:
				out.Type = p.symbolType(file, obj.Type)
				/* plantuml may need this, but data is incorrect
				item := &model.Field{
					Name: "type",
					Type: out.Type,
				}
				out.Fields = append(out.Fields, item)
				*/
			}
		default:
		}
	}
}

func (p *collector) parseStruct(structInfo *model.Declaration, file *ast.File, spec *ast.TypeSpec, obj *ast.StructType) {
	goPath := structInfo.Name

	if spec != nil && spec.TypeParams != nil {
		names := []string{}
		if spec != nil && spec.TypeParams != nil {
			for _, field := range spec.TypeParams.List { // loop over all TypeParam fields
				var constraint string
				switch t := field.Type.(type) {
				case *ast.Ident:
					constraint = t.Name
				case *ast.SelectorExpr:
					if x, ok := t.X.(*ast.Ident); ok {
						constraint = x.Name + "." + t.Sel.Name
					}
				default:
					constraint = "unknown"
				}

				// combine field names with constraint
				for _, ident := range field.Names { // loop over Names inside this Field
					names = append(names, ident.Name+" "+constraint)
				}
			}
		}

		structInfo.Arguments = names
	}

	for _, field := range obj.Fields.List {
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

		jsonName := JSONTagName(tagValue)
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

		v := &model.Field{
			Doc:     TrimSpace(field.Doc),
			Comment: TrimSpace(field.Comment),

			Name: goName,
			Path: fieldPath,
			Type: p.symbolType(file, field.Type),
			Tag:  tagValue,

			JSONName: jsonName,
		}

		structInfo.Fields = append(structInfo.Fields, v)
	}
	return
}
