package restore

import (
	"strings"
	"testing"

	"github.com/titpetric/tools/splint/model"
)

// definition is a package of two files, between them importing more than
// either declaration reaches.
func definition() *model.Definition {
	return &model.Definition{
		Package: model.Package{Package: "model", ImportPath: "example.com/model", Path: "."},
		Imports: model.StringSet{
			"trace.go": {`"time"`, `"context"`, `"encoding/json"`, `_ "embed"`},
			"auth.go":  {`"net/http"`, `"golang.org/x/crypto/bcrypt"`, `crypt "crypto/subtle"`},
		},
	}
}

// TestFileImports covers what a restored file imports: what its declarations
// reach and nothing else.
func TestFileImports(t *testing.T) {
	decls := model.DeclarationList{{
		Kind: model.TypeKind, Name: "Trace", File: "trace.go",
		Source: "// Trace is a recorded trace.\ntype Trace struct {\n\tStart time.Time\n}",
	}}

	body := "package model\n\n" + decls[0].Source + "\n"
	got, err := fileImports(definition(), decls, body)
	if err != nil {
		t.Fatalf("fileImports() error = %v", err)
	}

	// time is reached, context and encoding/json are not, and the blank
	// import of the file comes along because nothing reaches it by name.
	want := []string{`"time"`, `_ "embed"`}
	if !equal(got, want) {
		t.Errorf("fileImports() = %v, want %v", got, want)
	}
}

// TestFileImportsReadsSyntaxNotText covers the package named in a comment or
// in a string, which is not an import of anything.
func TestFileImportsReadsSyntaxNotText(t *testing.T) {
	decls := model.DeclarationList{{
		Kind: model.FuncKind, Name: "Render", File: "auth.go",
		Source: "// Render writes it out, the way json.Marshal would.\n" +
			"func Render() string {\n\treturn \"context.Background\"\n}",
	}}

	got, err := fileImports(definition(), decls, "package model\n\n"+decls[0].Source+"\n")
	if err != nil {
		t.Fatalf("fileImports() error = %v", err)
	}
	if len(got) != 0 {
		t.Errorf("fileImports() = %v, want none: it reaches nothing", got)
	}
}

// TestFileImportsAlias covers the import written under a name of its own,
// which is the name the body reaches it by.
func TestFileImportsAlias(t *testing.T) {
	decls := model.DeclarationList{{
		Kind: model.FuncKind, Name: "Equal", File: "auth.go",
		Source: "func Equal(a, b []byte) bool {\n\treturn crypt.ConstantTimeCompare(a, b) == 1\n}",
	}}

	got, err := fileImports(definition(), decls, "package model\n\n"+decls[0].Source+"\n")
	if err != nil {
		t.Fatalf("fileImports() error = %v", err)
	}
	if !equal(got, []string{`crypt "crypto/subtle"`}) {
		t.Errorf("fileImports() = %v, want the aliased import", got)
	}
}

// TestFileImportsCollision covers two files of the source package reaching
// different packages under one name. Merging them into one file is a file that
// does not compile, so it is reported instead.
func TestFileImportsCollision(t *testing.T) {
	def := definition()
	def.Imports["render.go"] = []string{`"text/template"`}
	def.Imports["page.go"] = []string{`"html/template"`}

	decls := model.DeclarationList{
		{Kind: model.FuncKind, Name: "Render", File: "render.go",
			Source: "func Render() *template.Template { return nil }"},
		{Kind: model.FuncKind, Name: "Page", File: "page.go",
			Source: "func Page() *template.Template { return nil }"},
	}

	body := "package model\n\n" + decls[0].Source + "\n\n" + decls[1].Source + "\n"
	_, err := fileImports(def, decls, body)
	if err == nil {
		t.Fatal("fileImports() accepted one name for two packages")
	}
	for _, want := range []string{"template", "text/template", "html/template"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("error = %v, want it to name %q", err, want)
		}
	}
}

func TestImportsOf(t *testing.T) {
	tests := map[string]string{
		`"net/http"`:                   "http",
		`"golang.org/x/crypto/bcrypt"`: "bcrypt",
		`crypt "crypto/subtle"`:        "crypt",
		`_ "embed"`:                    "_",
		`. "errors"`:                   ".",
		// A major version is not the name a package is reached by.
		`"github.com/go-chi/chi/v5"`: "chi",
	}

	for literal, want := range tests {
		got := importsOf([]string{literal})
		if len(got) != 1 || got[0].name != want {
			t.Errorf("importsOf(%q) = %#v, want the name %q", literal, got, want)
		}
	}
}

// equal reports two lists holding the same entries in the same order.
func equal(got, want []string) bool {
	if len(got) != len(want) {
		return false
	}
	for i := range got {
		if got[i] != want[i] {
			return false
		}
	}
	return true
}
