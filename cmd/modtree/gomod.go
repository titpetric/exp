package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/mod/modfile"
)

// parseGoWork returns relative paths listed under 'use' in go.work
func parseGoWork(file string) ([]string, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var dirs []string
	scanner := bufio.NewScanner(f)
	inUseBlock := false

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "//") {
			continue
		}
		if strings.HasPrefix(line, "use") && strings.HasSuffix(line, "(") {
			inUseBlock = true
			continue
		}
		if inUseBlock {
			if line == ")" {
				inUseBlock = false
				continue
			}
			dirs = append(dirs, line)
		}
	}
	return dirs, scanner.Err()
}

func readModulePath(dir string) (string, error) {
	data, err := os.ReadFile(filepath.Join(dir, "go.mod"))
	if err != nil {
		return "", err
	}

	mod, err := modfile.Parse("go.mod", data, nil)
	if err != nil {
		return "", err
	}

	if mod.Module == nil {
		return "", fmt.Errorf("no module declaration")
	}

	return mod.Module.Mod.Path, nil
}

func readRequiresVersioned(dir string) ([]requireInfo, error) {
	data, err := os.ReadFile(filepath.Join(dir, "go.mod"))
	if err != nil {
		return nil, err
	}

	mod, err := modfile.Parse("go.mod", data, nil)
	if err != nil {
		return nil, err
	}

	var reqs []requireInfo
	for _, r := range mod.Require {
		reqs = append(reqs, requireInfo{
			path:    r.Mod.Path,
			version: r.Mod.Version,
		})
	}
	return reqs, nil
}
