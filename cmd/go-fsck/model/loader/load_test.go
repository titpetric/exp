package loader_test

import (
	"testing"

	"github.com/kortschak/utter"
	"github.com/stretchr/testify/assert"

	"github.com/titpetric/exp/cmd/go-fsck/model"
	"github.com/titpetric/exp/cmd/go-fsck/model/loader"
)

func TestLoad(t *testing.T) {
	utter.Config.IgnoreUnexported = true
	utter.Config.OmitZero = true
	utter.Config.ElideType = true

	defs, err := loader.Load(&model.Package{
		Path: ".",
	}, true, true)

	assert.NoError(t, err)
	assert.NotNil(t, defs)

	utter.Dump(defs)
}
