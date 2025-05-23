package docs

import (
	"fmt"
	"go/ast"
	"strings"

	"github.com/titpetric/exp/cmd/go-fsck/model"
)

func renderPlantUML(_ *options, defs []*model.Definition) error {
	var links []string

	addLink := func(link string) {
		links = append(links, link)
	}

	fmt.Println("@startuml")
	fmt.Println("")

	allTypes := make(map[string]*model.Declaration)
	allFuncs := model.NewStringSet()

	for _, def := range defs {
		for _, t := range def.Types {
			allTypes[t.Name] = t
		}
	}

	for _, def := range defs {
		for _, t := range def.Funcs {
			receiver := t.ReceiverTypeRef()
			if receiver == "" {
				continue
			}

			if _, ok := allTypes[receiver]; ok {
				allFuncs.Add(receiver, t.Signature)
			}
		}
	}

	for _, def := range defs {
		if def.Package.TestPackage || strings.HasSuffix(def.Package.Package, "_test") {
			continue
		}

		for _, t := range def.Types {
			if len(t.Fields) == 0 {
				continue
			}

			for _, name := range t.Names {
				fmt.Println("class", name, "{")
				for _, f := range t.Fields {
					typeRef := f.TypeRef()

					if f.Name == "" {
						addLink(fmt.Sprintf("%s --|> %s : embeds", name, typeRef))
						continue
					}

					if strings.HasPrefix(f.Type, "struct") {
						f.Type = "struct"
					}
					if strings.HasPrefix(f.Type, "interface") {
						f.Type = "interface"
					}

					if ast.IsExported(f.Name) {
						fmt.Println("  +", f.Name+":", f.Type)
					} else {
						fmt.Println("  -", f.Name+":", f.Type)
					}
					if _, ok := allTypes[typeRef]; ok {
						addLink(fmt.Sprintf("%s --> %s : %s", name, typeRef, f.Name))
					}
				}
			}

			name := t.Name

			fmt.Println("class", name, "{")
			for _, f := range t.Fields {
				typeRef := f.TypeRef()

				if f.Name == "" {
					addLink(fmt.Sprintf("%s --|> %s : embeded by", typeRef, name))
					continue
				}

				if strings.HasPrefix(f.Type, "struct") {
					f.Type = "struct"
				}
				if strings.HasPrefix(f.Type, "interface") {
					f.Type = "interface"
				}

				if ast.IsExported(f.Name) {
					fmt.Println("  +", f.Name+":", f.Type)
				} else {
					fmt.Println("  -", f.Name+":", f.Type)
				}

				if _, ok := allTypes[typeRef]; ok {
					addLink(fmt.Sprintf("%s --> %s : %s", name, typeRef, f.Name))
				}
			}

			if t.Doc != "" {
				addLink("")
				addLink("note top of " + name)
				addLink(t.Doc)
				addLink("end note")
				addLink("")
			}

			addLink("")

			funcList := allFuncs.Get(t.Name)
			for _, sig := range funcList {
				funcName := strings.SplitN(sig, " ", 2)[0]
				funcInfo := def.Funcs.Find(func(d *model.Declaration) bool {
					return d.Receiver == t.Name && d.Name == funcName
				})
				if funcInfo == nil {
					continue
				}

				for _, argType := range funcInfo.Returns {
					cleanType := model.TypeRef(argType)
					if _, ok := allTypes[cleanType]; ok {
						addLink(fmt.Sprintf("%s --> %s : %s", name, cleanType, funcName+"()"))
						break
					}
				}

				if ast.IsExported(sig) {
					fmt.Println("  +", sig)
				} else {
					// fmt.Println("  -", sig)
				}
			}

			fmt.Println("}")
			fmt.Println()
		}
	}

	for _, link := range links {
		fmt.Println(link)
	}
	if len(links) > 0 {
		fmt.Println()
	}

	fmt.Println("@enduml")

	return nil
}
