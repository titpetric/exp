package generic

import (
	"embed"
	"fmt"
	"html/template"
	"path/filepath"
)

func templateLoaderEmbed(files *embed.FS) func(string, string, template.FuncMap) (*template.Template, error) {
	return func(filename string, commonTemplate string, funcs template.FuncMap) (*template.Template, error) {
		return template.New(filepath.Base(filename)).Funcs(funcs).ParseFS(files, filename, commonTemplate)
	}
}

func templateLoaderFile(filename string, commonTemplate string, funcs template.FuncMap) (*template.Template, error) {
	return template.New(filepath.Base(filename)).Funcs(funcs).ParseFiles(filename, commonTemplate)
}

// templateLoader will load the requested template + the common template (_common.tpl).
// This allows reuse of template blocks from the common template in all templates.
func templateLoader(name string, commonTemplate string, files *embed.FS, funcs template.FuncMap) (*template.Template, error) {
	var (
		loadFromEmbed = templateLoaderEmbed(files)
		loadFromFile  = templateLoaderFile
	)

	// loaders prefer local files, enabling override of the embedded ones.
	loaders := []func(string, string, template.FuncMap) (*template.Template, error){
		loadFromFile,
		loadFromEmbed,
	}

	var lastErr error
	for _, loader := range loaders {
		t, err := loader(name, commonTemplate, funcs)
		if err == nil {
			return t, nil
		}
		lastErr = err
	}

	return nil, fmt.Errorf("no such template: %s: %w", name, lastErr)
}
