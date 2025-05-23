package example_test

import (
	"fmt"
	"go/parser"
	"go/printer"
	"go/token"
	"io"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/titpetric/exp/cmd/go-fsck/internal/ast"
)

func PrintSource(out io.Writer, fset *token.FileSet, node any) error {
	return printer.Fprint(out, fset, node)
}

func TestPrinter(t *testing.T) {
	fset := token.NewFileSet()
	fs, err := parser.ParseDir(fset, ".", nil, parser.ParseComments)
	assert.NoError(t, err)

	for _, pkg := range fs {
		for _, f := range pkg.Files {
			for _, decl := range f.Decls {
				var out strings.Builder
				x := &printer.CommentedNode{
					Node:     decl,
					Comments: f.Comments,
				}
				assert.NoError(t, ast.PrintSource(&out, fset, x))
				fmt.Println(out.String())
			}
		}
	}
}
