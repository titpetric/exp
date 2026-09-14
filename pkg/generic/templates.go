package generic

import (
	"embed"
	"encoding/json"
	"html/template"
	"io"
)

type TemplateRenderer[T any] func(w io.Writer, templateName string, data T) error

type TemplateFuncMap interface {
	FuncMap() template.FuncMap
}

// NewTempleRenderer provides a type safe callback to render a template in a type-safe manner.
// The passed type needs to implement a FuncMap to provide it's own template APIs if needed.
// It's common to provide these to print time.Time values in human readable formats (localisation).
// The only function provided so far is \`json\`, allowing developers to inspect template data
// structures from the browser. Depending on the context of html/template, this output may
// be escaped, to sanitize it for HTML.
func NewTemplateRenderer[T TemplateFuncMap](files *embed.FS, defaultTemplate string) TemplateRenderer[T] {
	// default functions for the templates
	templateFuncs := template.FuncMap{
		"json": func(in interface{}) (string, error) {
			out, err := json.MarshalIndent(in, "", "  ")
			return string(out), err
		},
	}

	return func(w io.Writer, templateName string, data T) error {
		fns := data.FuncMap()
		for key, val := range fns {
			templateFuncs[key] = val
		}

		t, err := templateLoader(templateName, defaultTemplate, files, templateFuncs)
		if err != nil {
			return err
		}

		return t.Execute(w, data)
	}
}
