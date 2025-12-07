package internal

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/mod/modfile"

	"github.com/titpetric/exp/cmd/go-fsck/internal/telemetry"
)

type Module struct {
	Filename   string
	Dir        string
	Path       string
	ImportPath string

	Valid bool
	Error error
}

func (m Module) String() string {
	if m.Error != nil {
		return fmt.Sprintf("%v (dir: %v, valid: %v, error: %v)", m.ImportPath, m.Dir, m.Valid, m.Error)
	}
	return fmt.Sprintf("%v (dir: %v, valid: %v)", m.ImportPath, m.Dir, m.Valid)
}

// ListModules finds all go.mod files under root and returns a slice of Modules.
func ListModules(root string, pattern string) ([]Module, error) {
	span := telemetry.Start("internal.ListModules")
	defer span.End()

	if root == "" {
		root = "."
	}

	absRoot, err := filepath.Abs(root)
	if err != nil {
		return nil, err
	}

	var modules []Module
	err = filepath.WalkDir(absRoot, func(filename string, d os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.Name() != "go.mod" {
			return nil
		}
		if pattern == "./..." || (pattern == "." && len(modules) == 0) {
			modules = append(modules, parseGoMod(filename, absRoot))
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return modules, nil
}

// parseGoMod reads the go.mod file using x/mod/modfile
func parseGoMod(filename string, rootPath string) (result Module) {
	dir := filepath.Dir(filename)
	cleanDir := strings.TrimPrefix(dir, rootPath)

	result.Filename = filename
	result.Dir = dir
	result.Path = cleanDir

	data, err := ioutil.ReadFile(filename)
	if err != nil {
		result.Error = err
		return
	}

	f, err := modfile.Parse(filename, data, nil)
	if err != nil {
		result.Error = err
		return
	}

	if f.Module == nil || f.Module.Mod.Path == "" {
		result.Error = errors.New("module declaration not found")
		return
	}

	importPath := f.Module.Mod.Path
	result.ImportPath = importPath

	if !strings.Contains(importPath, ".") {
		result.Error = errors.New("module not importable")
		return
	}

	if cleanDir != "" && !strings.HasSuffix(importPath, cleanDir) {
		result.Error = fmt.Errorf("invalid import path: %s, want suffix: %s", importPath, cleanDir)
		return
	}

	result.Valid = true
	return
}
