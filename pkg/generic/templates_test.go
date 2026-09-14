package generic_test

import (
	"bytes"
	"embed"
	"html"
	"html/template"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/titpetric/exp/pkg/generic"
)

//go:embed testdata/templates/*
var templatesFS embed.FS

// Page implements TemplateFuncMap used by NewTemplateRenderer
type Page struct {
	Title string
	Body  string
}

func (p Page) FuncMap() template.FuncMap {
	return template.FuncMap{
		"greet": func(s string) string { return "Hello " + s },
	}
}

func TestTemplateRenderer(t *testing.T) {
	assert := assert.New(t)

	render := generic.NewTemplateRenderer[Page](nil, "testdata/templates/_common.tpl")

	page := Page{Title: "Test Title", Body: "Test Body"}
	var buf bytes.Buffer

	err := render(&buf, "testdata/templates/index.tpl", page)
	assert.NoError(err, "render should not return an error")

	out := buf.String()

	// basic checks: header, body, custom func output, and json output
	assert.Contains(out, "<h1>Test Title</h1>", "header should be rendered from _common.tpl")
	assert.Contains(out, "<div>Test Body</div>", "body should be rendered from index.tpl")
	assert.Contains(out, "Hello Test Title", "custom template function 'greet' should be available")

	// html/template escapes quotes, so unwrap HTML entities before checking the JSON payload
	unescaped := html.UnescapeString(out)
	assert.Contains(unescaped, `"Title": "Test Title"`, "json helper should include the Title field in output when unescaped")
}
