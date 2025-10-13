package collector

import (
	"go/ast"
	"strings"

	"github.com/fzipp/gocyclo"
	"github.com/uudashr/gocognit"

	"github.com/titpetric/exp/cmd/go-fsck/model"
)

func complexity(in *ast.FuncDecl, s string) *model.Complexity {
	lines := strings.Count(s, "\n")
	if len(s) > 0 && !strings.HasSuffix(s, "\n") {
		lines++
	}

	return &model.Complexity{
		Cognitive:  gocognit.Complexity(in),
		Cyclomatic: gocyclo.Complexity(in),
		Lines:      lines,
	}
}
