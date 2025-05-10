package docs

import (
	"fmt"
	"go/ast"
	"strings"

	"golang.org/x/exp/slices"

	"github.com/TykTechnologies/exp/cmd/go-fsck/model"
)

func renderPlantUML(defs []*model.Definition) error {
	var links []string

	addLink := func(link string) {
		if !slices.Contains(links, link) {
			links = append(links, link)
		}
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

			fmt.Println("class", t.Name, "{")
			for _, f := range t.Fields {
				typeRef := f.TypeRef()

				if f.Name == "" {
					addLink(fmt.Sprintf("%s --|> %s : embeds", t.Name, typeRef))
					continue
				}

				if ast.IsExported(f.Name) {
					fmt.Println("  +", f.Name+":", f.Type)
				} else {
					fmt.Println("  -", f.Name+":", f.Type)
				}
				if _, ok := allTypes[typeRef]; ok {
					addLink(fmt.Sprintf("%s --> %s : types", t.Name, typeRef))
				}
			}

			funcList := allFuncs.Get(t.Name)
			for _, sig := range funcList {
				if ast.IsExported(sig) {
					fmt.Println("  +", sig)
				} else {
					fmt.Println("  -", sig)
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
