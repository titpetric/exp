//go:build eyeball
// +build eyeball

package model_test

import (
	"testing"

	"github.com/davecgh/go-spew/spew"

	"github.com/titpetric/exp/cmd/schema-gen/model"
)

func Test_DefinitionsList_Sort(t *testing.T) {
	s, _ := model.Load("model.json")
	spew.Dump(s.Declarations.GetOrder("PackageInfo"))
}
