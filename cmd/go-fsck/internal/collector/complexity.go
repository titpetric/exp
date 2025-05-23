package collector

import (
	"go/ast"

	"github.com/fzipp/gocyclo"
	"github.com/uudashr/gocognit"

	"github.com/titpetric/exp/cmd/go-fsck/model"
)

func complexity(in *ast.FuncDecl) *model.Complexity {
	return &model.Complexity{
		Cognitive:  gocognit.Complexity(in),
		Cyclomatic: gocyclo.Complexity(in),
	}
}
