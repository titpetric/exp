package ast_test

import (
	"go/parser"
	"go/token"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/titpetric/exp/cmd/go-fsck/internal/ast"
)

const src = `package example

// Global func comment
func GlobalFunc() error {
	// holds the error
	var err error	// the err var

	// inline comment
	err = nil

	return err
}`

func TestPrint(t *testing.T) {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "", src, parser.ParseComments)
	assert.NoError(t, err)

	var out strings.Builder
	assert.NoError(t, ast.PrintSource(&out, fset, ast.CommentedNode(f, f)))

	assert.Equal(t, src, strings.TrimSpace(out.String()))
}
