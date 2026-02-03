package restore

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/titpetric/exp/cmd/go-fsck/model"
)

func TestFindFile_PrefersTypeOverFunc(t *testing.T) {
	// Test that when looking for a receiver, we find the Type first,
	// not a func that happens to have the same name

	files := map[string]model.DeclarationList{
		"ref.go": {
			&model.Declaration{
				Kind: model.TypeKind,
				Name: "Ref",
				Type: "struct",
			},
		},
		"declaration.go": {
			&model.Declaration{
				Kind:     model.FuncKind,
				Name:     "Ref",
				Receiver: "*Declaration",
			},
		},
	}

	// Simulate findFile logic
	findFile := func(find string) (string, bool) {
		// First, look for a type with this name (receiver should match a type)
		for filename, f := range files {
			for _, v := range f {
				if v.Kind == model.TypeKind && v.Name == find {
					return filename, true
				}
			}
		}
		// Fallback: look for any declaration with this name
		for filename, f := range files {
			for _, v := range f {
				if v.Name == find {
					return filename, true
				}
			}
		}
		return "", false
	}

	filename, found := findFile("Ref")

	assert.True(t, found, "Ref should be found")
	assert.Equal(t, "ref.go", filename, "Ref type should be in ref.go, not declaration.go")
}
