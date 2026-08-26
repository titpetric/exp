package internal

import (
	"os"
	"path/filepath"
	"sort"

	"golang.org/x/mod/modfile"

	"github.com/titpetric/exp/cmd/go-fsck/model"
)

// ReadModule parses the go.mod at filename into the module facts the model
// carries.
//
// ListModules parses the same file, but only far enough to decide whether the
// module is importable. What a release comparison needs is the rest of it: the
// go and toolchain directives, and every requirement with the version it is
// pinned to.
func ReadModule(filename string) (*model.Module, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	parsed, err := modfile.Parse(filename, data, nil)
	if err != nil {
		return nil, err
	}

	return moduleFrom(parsed), nil
}

// FindModule returns the module governing dir, found by walking up from it
// until a go.mod turns up.
//
// Extraction is pointed at a package directory as often as at a module root,
// and `go-fsck extract -i model/` has to reach the go.mod a level or more
// above it. A tree holding no go.mod at all is not an error: it still has
// packages worth reporting, and they are reported without a module.
func FindModule(dir string) (*model.Module, error) {
	current, err := filepath.Abs(dir)
	if err != nil {
		return nil, err
	}

	for {
		filename := filepath.Join(current, "go.mod")
		if _, err := os.Stat(filename); err == nil {
			return ReadModule(filename)
		}

		parent := filepath.Dir(current)
		if parent == current {
			return nil, nil
		}
		current = parent
	}
}

// moduleFrom copies a parsed go.mod into the model.
//
// The require block is written in whatever order go mod tidy last left it, and
// a module is free to hold several of them, so the lists are sorted by module
// path. Two revisions that require the same things then produce the same
// model, and a diff of the two reports nothing.
func moduleFrom(parsed *modfile.File) *model.Module {
	result := &model.Module{}

	if parsed.Module != nil {
		result.Path = parsed.Module.Mod.Path
	}
	if parsed.Go != nil {
		result.GoVersion = parsed.Go.Version
	}
	if parsed.Toolchain != nil {
		result.Toolchain = parsed.Toolchain.Name
	}

	for _, require := range parsed.Require {
		result.Requires = append(result.Requires, model.Require{
			Path:     require.Mod.Path,
			Version:  require.Mod.Version,
			Indirect: require.Indirect,
		})
	}
	sort.Slice(result.Requires, func(i, j int) bool {
		return result.Requires[i].Path < result.Requires[j].Path
	})

	for _, replace := range parsed.Replace {
		result.Replaces = append(result.Replaces, model.Replace{
			Path:       replace.Old.Path,
			Version:    replace.Old.Version,
			NewPath:    replace.New.Path,
			NewVersion: replace.New.Version,
		})
	}
	sort.Slice(result.Replaces, func(i, j int) bool {
		if result.Replaces[i].Path != result.Replaces[j].Path {
			return result.Replaces[i].Path < result.Replaces[j].Path
		}
		return result.Replaces[i].Version < result.Replaces[j].Version
	})

	return result
}
