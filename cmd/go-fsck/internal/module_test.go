package internal_test

import (
	"testing"

	"github.com/titpetric/exp/cmd/go-fsck/internal"
)

// TestLoadModules will traverse the folder structure for go.mod files;
// It reads the go module and matches it against the relative path of
// the current working directory.
//
// Use: `go test -c module_test.go` and then `./internal.test -test.v` from any source tree.
func TestLoadModules(t *testing.T) {
	mods, err := internal.ListModules(".")
	if err != nil {
		t.Fatalf("ListModules returned error: %v", err)
	}

	if len(mods) == 0 {
		t.Log("no modules found")
		return
	}

	for _, m := range mods {
		t.Logf("module: %s", m)
	}
}
