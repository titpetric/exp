package docs

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/titpetric/exp/cmd/go-fsck/model"
)

func TestIsExample(t *testing.T) {
	tests := []struct {
		name     string
		decl     *model.Declaration
		expected bool
	}{
		{
			name:     "package example",
			decl:     &model.Declaration{Kind: model.FuncKind, File: "example_test.go", Name: "Example"},
			expected: true,
		},
		{
			name:     "symbol example",
			decl:     &model.Declaration{Kind: model.FuncKind, File: "example_test.go", Name: "ExampleManager"},
			expected: true,
		},
		{
			name:     "method example with suffix",
			decl:     &model.Declaration{Kind: model.FuncKind, File: "example_test.go", Name: "ExampleManager_Apply_failure"},
			expected: true,
		},
		{
			name:     "function starting with the word",
			decl:     &model.Declaration{Kind: model.FuncKind, File: "example_test.go", Name: "Examples"},
			expected: false,
		},
		{
			name:     "outside a test file",
			decl:     &model.Declaration{Kind: model.FuncKind, File: "example.go", Name: "ExampleManager"},
			expected: false,
		},
		{
			name:     "not a function",
			decl:     &model.Declaration{Kind: model.VarKind, File: "example_test.go", Name: "ExampleManager"},
			expected: false,
		},
		{
			name: "takes arguments",
			decl: &model.Declaration{
				Kind:      model.FuncKind,
				File:      "example_test.go",
				Name:      "ExampleManager",
				Arguments: []string{"t *testing.T"},
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.expected, isExample(tt.decl))
		})
	}
}

func TestCollectExamples(t *testing.T) {
	defs := []*model.Definition{
		{
			Package: model.Package{Package: "migrate", ImportPath: "github.com/user/migrate"},
			Funcs: model.DeclarationList{
				{Kind: model.FuncKind, File: "apply.go", Name: "Apply"},
			},
		},
		{
			Package: model.Package{
				Package:     "migrate_test",
				ImportPath:  "github.com/user/migrate_test",
				TestPackage: true,
			},
			Funcs: model.DeclarationList{
				{Kind: model.FuncKind, File: "example_test.go", Name: "ExampleManager_Load"},
				{Kind: model.FuncKind, File: "example_test.go", Name: "ExampleManager_Apply"},
				{Kind: model.FuncKind, File: "apply_test.go", Name: "TestApply"},
			},
		},
	}

	examples := collectExamples(defs)

	require.Len(t, examples, 1, "examples are keyed by the package they document")

	names := []string{}
	for _, fn := range examples["github.com/user/migrate"] {
		names = append(names, fn.Name)
	}
	require.Equal(t, []string{"ExampleManager_Apply", "ExampleManager_Load"}, names)
}

func TestExampleSource(t *testing.T) {
	decl := &model.Declaration{
		Doc:    "ExampleManager applies a migration.",
		Source: "// ExampleManager applies a migration.\nfunc ExampleManager() {\n\tfmt.Println(\"ok\")\n}",
	}

	require.Equal(t, "func ExampleManager() {\n\tfmt.Println(\"ok\")\n}", exampleSource(decl))

	require.Empty(t, exampleSource(&model.Declaration{}), "a model without sources has no example to print")
}

func TestRenderExamples(t *testing.T) {
	items := model.DeclarationList{
		{
			Name:   "ExampleManager",
			Doc:    "ExampleManager applies a migration.",
			Source: "// ExampleManager applies a migration.\nfunc ExampleManager() {\n\tfmt.Println(\"ok\")\n}",
		},
	}

	out := renderExamples(items)

	require.Contains(t, out, "## Examples\n\n")
	require.Contains(t, out, "<section name=\"ExampleManager\">\n\n")
	require.Contains(t, out, "### ExampleManager\n\n")
	require.Contains(t, out, "ExampleManager applies a migration.\n\n")
	require.Contains(t, out, "```go\nfunc ExampleManager() {\n\tfmt.Println(\"ok\")\n}\n```\n\n")
	require.True(t, strings.HasSuffix(out, "</section>\n\n"))

	require.Empty(t, renderExamples(nil), "no examples is no heading")
	require.Empty(t,
		renderExamples(model.DeclarationList{{Name: "ExampleManager"}}),
		"an example with no source is left out, and it was the only one",
	)
}
