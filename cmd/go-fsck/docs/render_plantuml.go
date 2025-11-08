package docs

import (
	"fmt"
	"go/ast"
	"strings"

	"github.com/titpetric/exp/cmd/go-fsck/model"
)

func renderPlantUML(opt *options, defs []*model.Definition) error {
	var links []string

	addLink := func(link string) {
		links = append(links, link)
	}

	fmt.Println("@startuml")
	fmt.Println("")

	allTypes := make(map[string]*model.Declaration)
	allPackages := make(map[string]*model.Package)
	allFuncs := make(map[string][]*model.Declaration)

	for _, def := range defs {
		allPackages[def.Package.ImportPath] = &def.Package
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

			receiver = def.Package.Namespace(".") + receiver

			allFuncs[receiver] = append(allFuncs[receiver], t)
		}
	}

	for _, def := range defs {
		importMap, _ := def.Imports.Map(def.Imports.All())

		lookup := func(name string) (*model.Package, bool) {
			importpath, ok := importMap[name]
			if !ok {
				return nil, false
			}

			pkg, ok := allPackages[importpath]
			return pkg, ok
		}

		if def.Package.TestPackage || strings.HasSuffix(def.Package.Package, "_test") {
			continue
		}

		namespace := def.Package.Namespace(".")

		for _, t := range def.Types {
			if len(t.Fields) == 0 {
				continue
			}

			for _, name := range t.GetNames() {
				var token = "class"
				if strings.HasPrefix(t.Type, "interface") {
					token = "interface"
				}

				for _, f := range t.Fields {
					if strings.HasPrefix(f.Type, "func") {
						token = "interface"
					}
				}

				if len(t.Arguments) > 0 {
					names := []string{}
					for _, name := range t.Arguments {
						names = append(names, name)
					}
					t.Arguments = names
					name += "[" + strings.Join(names, ", ") + "]"
				}

				fmt.Println(token, fmt.Sprintf("%q", namespace+name), "{")
				for _, f := range t.Fields {
					typeRef := f.TypeRef()

					if f.Name == "" {
						if strings.Contains(typeRef, ".") {
							parts := strings.SplitN(typeRef, ".", 2)
							packageName, typeName := parts[0], parts[1]
							if p, ok := lookup(packageName); ok {
								addLink(fmt.Sprintf("%q --|> %q : embeds", namespace+name, p.Namespace(".")+typeName))
							}
						} else {
							addLink(fmt.Sprintf("%q --|> %q : embeds", namespace+name, namespace+typeRef))
						}
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

					if token != "interface" {
						if strings.Contains(typeRef, ".") {
							parts := strings.SplitN(typeRef, ".", 2)
							packageName, typeName := parts[0], parts[1]
							if p, ok := lookup(packageName); ok {
								addLink(fmt.Sprintf("%q --> %q : .%s", namespace+name, p.Namespace(".")+typeName, f.Name))
							}
						} else {
							if _, ok := model.ToType(typeRef); ok {
								addLink(fmt.Sprintf("%q --> %q : .%s", namespace+name, namespace+typeRef, f.Name))
							}
						}
					}
				}
				if opt.docs && t.Doc != "" {
					addLink("")
					addLink("note top of " + namespace + name)
					addLink(t.Doc)
					addLink("end note")
					addLink("")
				}

				addLink("")

				if token == "interface" {
					continue
				}

				funcList := allFuncs[namespace+t.Name]
				for _, sig := range funcList {
					funcInfo := sig
					funcName := sig.Name

					func() {
						for _, argType := range funcInfo.Returns {
							typeRef := model.TypeRef(argType)

							if strings.Contains(typeRef, ".") {
								parts := strings.SplitN(typeRef, ".", 2)
								packageName, typeName := parts[0], parts[1]
								if p, ok := lookup(packageName); ok {
									addLink(fmt.Sprintf("%q --> %q : .%s()", namespace+name, p.Namespace(".")+typeName, funcName))
									return
								}
							}
						}
						for _, argType := range funcInfo.Arguments {
							typeRef := model.TypeRef(argType)

							if strings.Contains(typeRef, ".") {
								parts := strings.SplitN(typeRef, ".", 2)
								packageName, typeName := parts[0], parts[1]
								if p, ok := lookup(packageName); ok {
									addLink(fmt.Sprintf("%q --> %q : .%s()", namespace+name, p.Namespace(".")+typeName, funcName))
									return
								}
							}
						}
					}()

					if ast.IsExported(sig.Signature) {
						fmt.Println("  +", sig.Signature)
					} else {
						// fmt.Println("  -", sig)
					}
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
