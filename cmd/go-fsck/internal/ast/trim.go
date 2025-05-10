package ast

import (
	"go/ast"
	"strings"
)

func TrimSpace(in *ast.CommentGroup) string {
	if in != nil {
		return strings.TrimSpace(in.Text())
	}
	return ""
}
