package rules

import (
	"testing"

	"github.com/titpetric/exp/cmd/go-fsck/model"
)

func TestFuncArgsLinter_ContextFirst(t *testing.T) {
	linter := NewFuncArgsLinter()
	defs := []*model.Definition{
		{
			Package: model.Package{Path: "test"},
			Funcs: []*model.Declaration{
				{
					Name:      "Do",
					Kind:      model.FuncKind,
					Arguments: []string{"context.Context", "string"},
					File:      "test.go",
					Line:      1,
				},
			},
		},
	}
	linter.Lint(defs)
	if len(linter.Issues()) != 0 {
		t.Errorf("expected 0 issues for context.Context first, got %d", len(linter.Issues()))
	}
}

func TestFuncArgsLinter_ContextNotFirst(t *testing.T) {
	linter := NewFuncArgsLinter()
	defs := []*model.Definition{
		{
			Package: model.Package{Path: "test"},
			Funcs: []*model.Declaration{
				{
					Name:      "Render",
					Kind:      model.FuncKind,
					Arguments: []string{"int", "AnyContext"},
					File:      "test.go",
					Line:      1,
				},
			},
		},
	}
	linter.Lint(defs)
	if len(linter.Issues()) != 1 {
		t.Errorf("expected 1 issue for context not first, got %d", len(linter.Issues()))
	}
}

func TestFuncArgsLinter_TimeDurationLast(t *testing.T) {
	linter := NewFuncArgsLinter()
	defs := []*model.Definition{
		{
			Package: model.Package{Path: "test"},
			Funcs: []*model.Declaration{
				{
					Name:      "Set",
					Kind:      model.FuncKind,
					Arguments: []string{"string", "time.Duration"},
					File:      "test.go",
					Line:      1,
				},
			},
		},
	}
	linter.Lint(defs)
	if len(linter.Issues()) != 0 {
		t.Errorf("expected 0 issues for time.Duration last, got %d", len(linter.Issues()))
	}
}

func TestFuncArgsLinter_TimeDurationNotLast(t *testing.T) {
	linter := NewFuncArgsLinter()
	defs := []*model.Definition{
		{
			Package: model.Package{Path: "test"},
			Funcs: []*model.Declaration{
				{
					Name:      "Set",
					Kind:      model.FuncKind,
					Arguments: []string{"time.Duration", "string"},
					File:      "test.go",
					Line:      1,
				},
			},
		},
	}
	linter.Lint(defs)
	if len(linter.Issues()) != 1 {
		t.Errorf("expected 1 issue for time.Duration not last, got %d", len(linter.Issues()))
	}
}

func TestFuncArgsLinter_DuplicateTypes(t *testing.T) {
	linter := NewFuncArgsLinter()
	defs := []*model.Definition{
		{
			Package: model.Package{Path: "test"},
			Funcs: []*model.Declaration{
				{
					Name:      "Equal",
					Kind:      model.FuncKind,
					Arguments: []string{"*T", "*T"},
					File:      "test.go",
					Line:      1,
				},
			},
		},
	}
	linter.Lint(defs)
	if len(linter.Issues()) != 1 {
		t.Errorf("expected 1 issue for duplicate types, got %d", len(linter.Issues()))
	}
}

func TestFuncArgsLinter_StringAnyAmbiguous(t *testing.T) {
	linter := NewFuncArgsLinter()
	defs := []*model.Definition{
		{
			Package: model.Package{Path: "test"},
			Funcs: []*model.Declaration{
				{
					Name:      "Fetch",
					Kind:      model.FuncKind,
					Arguments: []string{"string", "map[string]any"},
					File:      "test.go",
					Line:      1,
				},
			},
		},
	}
	linter.Lint(defs)
	if len(linter.Issues()) != 0 {
		t.Errorf("expected 0 issues for (string, any) ambiguous, got %d", len(linter.Issues()))
	}
}

func TestFuncArgsLinter_VariadicArguments(t *testing.T) {
	linter := NewFuncArgsLinter()
	defs := []*model.Definition{
		{
			Package: model.Package{Path: "test"},
			Funcs: []*model.Declaration{
				{
					Name:      "Do",
					Kind:      model.FuncKind,
					Arguments: []string{"string", "...string"},
					File:      "test.go",
					Line:      1,
				},
			},
		},
	}
	linter.Lint(defs)
	if len(linter.Issues()) != 0 {
		t.Errorf("expected 0 issues for variadic arguments, got %d", len(linter.Issues()))
	}
}

func TestFuncArgsLinter_InterfaceBeforeStruct(t *testing.T) {
	linter := NewFuncArgsLinter()
	defs := []*model.Definition{
		{
			Package: model.Package{Path: "test"},
			Types: []*model.Declaration{
				{
					Name: "Reader",
					Type: "interface",
				},
			},
			Funcs: []*model.Declaration{
				{
					Name:      "Do",
					Kind:      model.FuncKind,
					Arguments: []string{"Reader", "*Config"},
					File:      "test.go",
					Line:      1,
				},
			},
		},
	}
	linter.Lint(defs)
	if len(linter.Issues()) != 0 {
		t.Errorf("expected 0 issues for interface before struct, got %d", len(linter.Issues()))
	}
}

func TestFuncArgsLinter_StructBeforeInterface(t *testing.T) {
	linter := NewFuncArgsLinter()
	defs := []*model.Definition{
		{
			Package: model.Package{Path: "test"},
			Types: []*model.Declaration{
				{
					Name: "Reader",
					Type: "interface",
				},
			},
			Funcs: []*model.Declaration{
				{
					Name:      "Do",
					Kind:      model.FuncKind,
					Arguments: []string{"*Config", "Reader"},
					File:      "test.go",
					Line:      1,
				},
			},
		},
	}
	linter.Lint(defs)
	if len(linter.Issues()) != 1 {
		t.Errorf("expected 1 issue for struct before interface, got %d", len(linter.Issues()))
	}
}

func TestFuncArgsLinter_SingleArgumentPasses(t *testing.T) {
	linter := NewFuncArgsLinter()
	defs := []*model.Definition{
		{
			Package: model.Package{Path: "test"},
			Funcs: []*model.Declaration{
				{
					Name:      "Process",
					Kind:      model.FuncKind,
					Arguments: []string{"context.Context"},
					File:      "test.go",
					Line:      1,
				},
			},
		},
	}
	linter.Lint(defs)
	if len(linter.Issues()) != 0 {
		t.Errorf("expected 0 issues for single argument, got %d", len(linter.Issues()))
	}
}

func TestFuncArgsLinter_Statistics(t *testing.T) {
	linter := NewFuncArgsLinter()
	defs := []*model.Definition{
		{
			Package: model.Package{Path: "test"},
			Funcs: []*model.Declaration{
				{Name: "A", Kind: model.FuncKind, Arguments: []string{}, File: "test.go", Line: 1},
				{Name: "B", Kind: model.FuncKind, Arguments: []string{"x"}, File: "test.go", Line: 2},
				{Name: "C", Kind: model.FuncKind, Arguments: []string{"x", "y"}, File: "test.go", Line: 3},
				{Name: "D", Kind: model.FuncKind, Arguments: []string{"x", "y", "z"}, File: "test.go", Line: 4},
			},
		},
	}
	linter.Lint(defs)
	summary := linter.IssueSummary()

	if summary["total_symbols"] != 4 {
		t.Errorf("expected total_symbols=4, got %v", summary["total_symbols"])
	}
	if summary["considered_funcs"] != 1 { // Only exactly 2 args
		t.Errorf("expected considered_funcs=1, got %v", summary["considered_funcs"])
	}
}

func TestFuncArgsLinter_MoreThanTwoArgumentsSkipped(t *testing.T) {
	linter := NewFuncArgsLinter()
	defs := []*model.Definition{
		{
			Package: model.Package{Path: "test"},
			Types: []*model.Declaration{
				{
					Name: "Reader",
					Type: "interface",
				},
			},
			Funcs: []*model.Declaration{
				{
					Name:      "Do",
					Kind:      model.FuncKind,
					Arguments: []string{"context.Context", "*Config", "Reader"},
					File:      "test.go",
					Line:      1,
				},
				{
					Name:      "Set",
					Kind:      model.FuncKind,
					Arguments: []string{"context.Context", "time.Duration", "string"},
					File:      "test.go",
					Line:      2,
				},
			},
		},
	}
	linter.Lint(defs)
	if len(linter.Issues()) != 0 {
		t.Errorf("expected 0 issues for functions with 3 arguments, got %d", len(linter.Issues()))
	}
	if got := linter.GetStatistics().ConsideredFuncs; got != 0 {
		t.Errorf("expected considered_funcs=0, got %d", got)
	}
}
