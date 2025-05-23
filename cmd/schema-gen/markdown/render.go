package markdown

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"

	"golang.org/x/exp/slices"

	"github.com/titpetric/exp/cmd/schema-gen/model"
)

func render(cfg *options) error {
	var (
		kind        = "markdown"
		inputFile   = cfg.inputFile
		outputFile  = cfg.outputFile
		packageName = cfg.packageName
		rootElement = cfg.rootElement
	)

	pkgInfos, err := model.Load(inputFile)
	if err != nil {
		return fmt.Errorf("Error loading package info for %s: %w", inputFile, err)
	}

	if packageName == "" {
		for _, pkgInfo := range pkgInfos {
			if strings.HasSuffix(pkgInfo.Name, "_test") {
				continue
			}
			packageName = pkgInfo.Name
			break
		}
	}

	for _, pkgInfo := range pkgInfos {
		// If we have multiple packages (tests), we only render one
		// of them, e.g. usually the public-facing API and not tests.
		// With the packageName option, this can be customized.
		if len(pkgInfos) != 1 && pkgInfo.Name != packageName {
			continue
		}

		order := pkgInfo.Declarations.GetOrder(rootElement)

		switch kind {
		case "markdown":
			body, err := renderMarkdown(cfg, packageName, sanitize(pkgInfo.Declarations), order)
			if err != nil {
				return err
			}
			if outputFile == "-" {
				_, err = os.Stdout.Write(body)
				return err
			}
			return os.WriteFile(outputFile, body, 0644)
		}
		return fmt.Errorf("Renderer %q not implemented", kind)
	}
	return fmt.Errorf("Uknown package name: %q", packageName)
}

func renderMarkdown(cfg *options, packageName string, schema model.DeclarationList, order []string) ([]byte, error) {
	output := new(bytes.Buffer)
	decls := schema.Find(order)

	allTypes := make([]string, 0, len(decls))
	for _, decl := range decls {
		allTypes = append(allTypes, decl.Name)
	}

	for _, decl := range decls {
		if err := renderMarkdownType(cfg, output, packageName, decl, allTypes); err != nil {
			return nil, err
		}
	}
	return output.Bytes(), nil
}

func renderMarkdownType(cfg *options, w io.Writer, packageName string, decl *model.TypeInfo, allTypes []string) error {
	canRender := func(name string) bool {
		if len(cfg.keep) > 0 {
			return slices.Contains(cfg.keep, name)
		}

		if len(cfg.skip) > 0 {
			return !slices.Contains(cfg.skip, name)
		}

		return true
	}

	if !canRender(decl.Name) {
		return nil
	}

	if decl.Name == cfg.rootElement && cfg.title != "" {
		// Override root element with title if provided
		fmt.Fprint(w, cfg.title+"\n\n")
	} else {
		name := decl.Name
		if cfg.full {
			name = packageName + "." + decl.Name
		}

		fmt.Fprintf(w, cfg.headingFormat+"\n\n", name)
	}

	if decl.Doc != "" {
		fmt.Fprintf(w, "%s\n\n", decl.Doc)
	}
	renderMarkdownFields(cfg, w, packageName+"."+decl.Name, decl, allTypes)
	return nil
}

func renderMarkdownFields(cfg *options, w io.Writer, packageName string, decl *model.TypeInfo, allTypes []string) {
	for _, field := range decl.Fields {
		jsonTag := strings.Split(field.JSONName, ",")

		sanitizedType := strings.TrimLeft(field.Type, "[]*")
		isKnown := slices.Contains(allTypes, sanitizedType)

		if isKnown {
			// Link the known type
			fmt.Fprintf(w, cfg.fieldFormat+"\n", jsonTag[0], field.Type, strings.ToLower(sanitizedType))

			// This prints the go field name as well.
			// fmt.Fprintf(w, "**Field: `%s` (%s, [%s](#%s))**\n", jsonTag[0], field.Name, field.Type, sanitizedType)
		} else {
			fieldType := fmt.Sprint(field.Type)

			if repl, ok := cfg.replace[fieldType]; ok {
				fieldType = repl
			}

			if strings.Contains(fieldType, ".") {
				fmt.Fprintf(os.Stderr, "WARN: field %s is not a JSON type (use --replace)\n", fieldType)
			}

			switch fieldType {
			case "bool":
				fieldType = "boolean"
			case "interface{}", "map[string]interface{}":
				fieldType = "any"
			}

			name := jsonTag[0]
			if cfg.full {
				name = packageName + "." + jsonTag[0]
			}

			fmt.Fprintf(w, cfg.fieldFormatKnown+"\n", name, fieldType)

			// This prints the go field name as well.
			// fmt.Fprintf(w, "**Field: `%s` (%s, `%s`)**\n", jsonTag[0], field.Name, field.Type)
		}
		if cfg.fieldSpacing {
			fmt.Fprintln(w)
		}

		if cfg.trim != "" {
			doclines := strings.Split(field.Doc, "\n")
			for _, v := range doclines {
				if strings.HasPrefix(v, cfg.trim) {
					fmt.Fprintln(w)
					fmt.Fprintf(w, "%s\n", docString(v))
					continue
				}
				fmt.Fprintf(w, "%s\n", v)
			}
			fmt.Fprintln(w)
		} else {
			fmt.Fprintf(w, "%s\n\n", field.Doc)
		}

		// This adds the line-level comment to the doc.
		// We don't need it but should likely put it behind a config option in the future.
		//	if field.Comment != "" {
		//		fmt.Fprintf(w, "> %s\n\n", field.Comment)
		//	}
	}

	if len(decl.Fields) == 0 {
		var (
			originalType  = decl.Type
			sanitizedType = decl.Type
			typeFormat    = "Type defined as `%s`"
		)
		switch {
		case strings.HasPrefix(sanitizedType, "[]"):
			sanitizedType = strings.TrimLeft(decl.Type, "[]*")
			originalType = sanitizedType
			typeFormat = "Type defined as array of `%s` values"
		case strings.HasPrefix(sanitizedType, "map["):
			typeFormat = "Type defined as object of `%s` values"
			sanitizedType = strings.Split(sanitizedType, "]")[1]
			if sanitizedType == "interface{}" {
				sanitizedType = "object"
			}
			originalType = sanitizedType
		}

		isKnown := slices.Contains(allTypes, sanitizedType)
		if isKnown {
			anchor := "#" + strings.ToLower(sanitizedType)
			fmt.Fprintf(w, typeFormat+", see [%s](%s) definition.\n\n", originalType, sanitizedType, anchor)
		}
	}
}

func docString(in string) string {
	out := strings.Trim(in, ".")
	return out + "."
}

func sanitize(x model.DeclarationList) model.DeclarationList {
	result := model.DeclarationList{}
	for _, decl := range x {
		// Move declaration doc comment into type decl if there
		// is only one and the comment is empty. Weird.
		if len(decl.Types) == 1 && decl.Types[0].Doc == "" {
			decl.Types[0].Doc = decl.Doc
			decl.Doc = ""
		}

		/*
			// Skip types with no exposed fields
			newTypes := model.TypeList{}
			for _, typeDecl := range decl.Types {
				if len(typeDecl.Fields) > 0 {
					newTypes = append(newTypes, typeDecl)
				}
			}
			decl.Types = newTypes

			if len(newTypes) > 0 {
				result = append(result, decl)
			}
		*/

		result = append(result, decl)
	}
	return result
}
