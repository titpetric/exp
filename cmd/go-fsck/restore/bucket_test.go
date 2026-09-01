package restore

import (
	"testing"

	"github.com/titpetric/tools/splint/model"
)

// declarations are a package of the shapes the layout has to place: a type
// with a method and a constructor, a typed const, an untyped var, an
// unexported helper, and the tests over them.
func declarations() model.DeclarationList {
	return model.DeclarationList{
		{Kind: model.TypeKind, Name: "Tracer", File: "tracer.go"},
		{Kind: model.TypeKind, Name: "SpanKind", File: "kind.go"},
		{Kind: model.FuncKind, Name: "Start", Receiver: "*Tracer", File: "tracer.go"},
		{Kind: model.FuncKind, Name: "NewTracer", File: "new.go"},
		{Kind: model.FuncKind, Name: "Mount", File: "mount.go"},
		{Kind: model.FuncKind, Name: "sanitise", File: "util.go"},
		{Kind: model.ConstKind, Name: "KindServer", Type: "SpanKind", File: "kind.go"},
		{Kind: model.ConstKind, Name: "DefaultPath", File: "options.go"},
		{Kind: model.VarKind, Name: "ErrNoTracer", File: "errors.go"},
		{Kind: model.VarKind, Name: "clock", File: "util.go"},
		{Kind: model.FuncKind, Name: "TestTracer", File: "tracer_test.go"},
		{Kind: model.FuncKind, Name: "TestTracer_Start", File: "tracer_test.go"},
		{Kind: model.FuncKind, Name: "TestMountRejectsNilRouter", File: "mount_test.go"},
		{Kind: model.FuncKind, Name: "TestEverythingTogether", File: "oida_test.go"},
	}
}

// TestLayout covers where each shape goes: a symbol in the file named for it,
// a method and a constructor beside the type, a typed const beside its type,
// and the unexported half in the file named for the package.
func TestLayout(t *testing.T) {
	layout := newLayout("oida", false, declarations())

	want := map[string]string{
		"Tracer":                    "tracer.go",
		"SpanKind":                  "span_kind.go",
		"Start":                     "tracer.go",
		"NewTracer":                 "tracer.go",
		"Mount":                     "mount.go",
		"sanitise":                  "oida.go",
		"KindServer":                "span_kind.go",
		"DefaultPath":               "const.go",
		"ErrNoTracer":               "vars.go",
		"clock":                     "oida.go",
		"TestTracer":                "tracer_test.go",
		"TestTracer_Start":          "tracer_test.go",
		"TestMountRejectsNilRouter": "mount_test.go",
		// A test the naming resolves to nothing tests the package.
		"TestEverythingTogether": "oida_test.go",
	}

	for _, decl := range declarations() {
		if got := layout.File(decl); got != want[decl.Name] {
			t.Errorf("%s goes in %s, want %s", decl.Name, got, want[decl.Name])
		}
	}
}

// TestLayoutSplit covers the flag: every symbol in the file named for it, the
// unexported ones included.
func TestLayoutSplit(t *testing.T) {
	layout := newLayout("oida", true, declarations())

	want := map[string]string{
		"sanitise":    "sanitise.go",
		"clock":       "clock.go",
		"DefaultPath": "default_path.go",
		"KindServer":  "kind_server.go",
		// A method and a constructor are part of the type, and stay with it.
		"Start":     "tracer.go",
		"NewTracer": "tracer.go",
		"Tracer":    "tracer.go",
	}

	for _, decl := range declarations() {
		file, wanted := want[decl.Name]
		if !wanted {
			continue
		}
		if got := layout.File(decl); got != file {
			t.Errorf("%s goes in %s, want %s", decl.Name, got, file)
		}
	}
}

// TestClaim covers the two packages of one directory: the tests inside a
// package and the tests beside it both collect what they do not name in a file
// named for the package, and one name cannot be two files.
func TestClaim(t *testing.T) {
	taken := map[string]string{}

	if got := claim(taken, "model_test.go", "model"); got != "model_test.go" {
		t.Errorf("the first to ask got %q", got)
	}
	if got := claim(taken, "model_test.go", "model"); got != "model_test.go" {
		t.Errorf("the same package asking again got %q", got)
	}
	if got := claim(taken, "model_test.go", "model_test"); got != "model_ext_test.go" {
		t.Errorf("the package beside it got %q, want a file of its own", got)
	}
}

func TestTestedName(t *testing.T) {
	tests := map[string]string{
		"TestTracer":         "Tracer",
		"TestTracer_Live":    "Tracer",
		"BenchmarkTracer":    "Tracer",
		"ExampleTracer_Live": "Tracer",
		"FuzzParse":          "Parse",
		"Tracer":             "Tracer",
	}

	for name, want := range tests {
		if got := testedName(name); got != want {
			t.Errorf("testedName(%q) = %q, want %q", name, got, want)
		}
	}
}

func TestToFilename(t *testing.T) {
	tests := map[string]string{
		"Tracer":      "tracer.go",
		"RouterFunc":  "router_func.go",
		"HTTPHandler": "http_handler.go",
		"OAuthToken":  "oauth_token.go",
		"ID":          "id.go",
	}

	for name, want := range tests {
		if got := toFilename(name); got != want {
			t.Errorf("toFilename(%q) = %q, want %q", name, got, want)
		}
	}
}
