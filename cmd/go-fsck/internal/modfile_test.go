package internal_test

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/titpetric/exp/cmd/go-fsck/internal"
	"github.com/titpetric/exp/cmd/go-fsck/model"
)

// writeModule writes a go.mod into a temporary directory and returns the
// directory, which is what the search walks up from.
func writeModule(t *testing.T, contents string) string {
	t.Helper()

	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte(contents), 0o644); err != nil {
		t.Fatalf("writing go.mod: %v", err)
	}
	return dir
}

func TestReadModule(t *testing.T) {
	dir := writeModule(t, `module example.com/x

go 1.24.0

toolchain go1.24.2

require (
	example.com/b v1.2.0
	example.com/a v0.4.0 // indirect
)

require example.com/c/v2 v2.0.0

replace example.com/b => example.com/fork/b v1.3.0

replace example.com/d v1.0.0 => ../d
`)

	got, err := internal.ReadModule(filepath.Join(dir, "go.mod"))
	if err != nil {
		t.Fatalf("ReadModule() error: %v", err)
	}

	want := &model.Module{
		Path:      "example.com/x",
		GoVersion: "1.24.0",
		Toolchain: "go1.24.2",
		Requires: []model.Require{
			{Path: "example.com/a", Version: "v0.4.0", Indirect: true},
			{Path: "example.com/b", Version: "v1.2.0"},
			{Path: "example.com/c/v2", Version: "v2.0.0"},
		},
		Replaces: []model.Replace{
			{Path: "example.com/b", NewPath: "example.com/fork/b", NewVersion: "v1.3.0"},
			{Path: "example.com/d", Version: "v1.0.0", NewPath: "../d"},
		},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("ReadModule() = %#v, want %#v", got, want)
	}
}

func TestFindModuleWalksUp(t *testing.T) {
	dir := writeModule(t, "module example.com/x\n\ngo 1.24.0\n")

	nested := filepath.Join(dir, "model", "loader")
	if err := os.MkdirAll(nested, 0o755); err != nil {
		t.Fatalf("creating nested directory: %v", err)
	}

	got, err := internal.FindModule(nested)
	if err != nil {
		t.Fatalf("FindModule() error: %v", err)
	}
	if got == nil || got.Path != "example.com/x" {
		t.Fatalf("FindModule() = %#v, want the module a level above", got)
	}
}

func TestFindModuleReportsNoModuleWithoutError(t *testing.T) {
	// A source tree with no go.mod still holds packages worth extracting, so
	// the absence is reported rather than raised.
	got, err := internal.FindModule(t.TempDir())
	if err != nil {
		t.Fatalf("FindModule() error: %v", err)
	}
	if got != nil {
		t.Fatalf("FindModule() = %#v, want none", got)
	}
}
