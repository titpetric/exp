package generic

import (
	"embed"
	"html/template"
	"net/http"
)

// TemplateExampleData is the model passed to the renderer's templates.
// It represents a user with exported string fields.
type TemplateExampleData struct {
	Name   string
	Link   string
	Avatar string
}

// FuncMap satisfies the TemplateFuncMap constraint required by NewTemplateRenderer.
func (TemplateExampleData) FuncMap() template.FuncMap {
	return template.FuncMap{}
}

// TemplateExample holds a TemplateRenderer specialized for TemplateExampleData.
type TemplateExample struct {
	view TemplateRenderer[TemplateExampleData]
}

// NewTemplateExample constructs a TemplateExample using the provided embedded FS.
func NewTemplateExample(files *embed.FS) *TemplateExample {
	return &TemplateExample{
		view: NewTemplateRenderer[TemplateExampleData](files, "testdata/templates/_common.tpl"),
	}
}

// ServeHTTP implements http.Handler. It constructs an TemplateExampleData instance
// and invokes the underlying TemplateRenderer to render the template.
func (t *TemplateExample) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	model := TemplateExampleData{
		Name:   "Alice Example",
		Link:   "https://example.com/alice",
		Avatar: "https://example.com/alice.png",
	}

	if err := t.view(w, "testdata/templates/example.tpl", model); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
