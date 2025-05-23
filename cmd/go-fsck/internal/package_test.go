package internal_test

import (
	"testing"

	"github.com/davecgh/go-spew/spew"
	"github.com/stretchr/testify/assert"

	"github.com/titpetric/exp/cmd/go-fsck/internal"
)

var ListPackages = internal.ListPackages

func TestListPackages(t *testing.T) {
	pkgs, err := ListPackages(".", "./...")
	assert.NoError(t, err)
	spew.Dump(pkgs)

	pkgs, err = ListPackages(".", ".")
	assert.NoError(t, err)
	spew.Dump(pkgs)
}
