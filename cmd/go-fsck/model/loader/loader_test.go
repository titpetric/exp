package loader_test

import (
	"testing"

	"github.com/kortschak/utter"
	"github.com/stretchr/testify/assert"

	"github.com/titpetric/exp/cmd/go-fsck/internal"
	"github.com/titpetric/exp/cmd/go-fsck/model/loader"
)

func TestLoad(t *testing.T) {
	utter.Config.IgnoreUnexported = true
	utter.Config.OmitZero = true
	utter.Config.ElideType = true

	// list current local packages
	packages, err := internal.ListPackages(".", ".")
	assert.NoError(t, err)
	assert.NotNil(t, packages)

	for _, p := range packages {
		defs, err := loader.Load(p, true, true)
		assert.NoError(t, err)
		assert.NotNil(t, defs)

		utter.Dump(defs)
	}
}
