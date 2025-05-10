package collector

import (
	"go/ast"
	"strings"

	. "github.com/TykTechnologies/exp/cmd/go-fsck/internal/ast"
	"github.com/TykTechnologies/exp/cmd/go-fsck/model"
)

func (p *collector) collectStructFields(out *model.Declaration, file *ast.File, decl *ast.GenDecl) {
	for _, spec := range decl.Specs {
		switch obj := spec.(type) {
		case *ast.TypeSpec:
			switch val := obj.Type.(type) {
			case *ast.StructType:
				p.parseStruct(out, file, val)
			default:
				out.Type = p.symbolType(file, obj.Type)
				item := &model.Field{
					Name: "type",
					Type: out.Type,
				}
				out.Fields = append(out.Fields, item)
			}
		default:
		}
	}
}

func (p *collector) parseStruct(structInfo *model.Declaration, file *ast.File, obj *ast.StructType) {
	var (
		goPath = structInfo.Name
	)

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
