package extract

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// defaultVersion is what a path with no version after it is read at.
const defaultVersion = "latest"

// isImportPath reports a source path that names a package of another module
// rather than a directory of this one.
//
// A directory that exists is a directory. What is left is read as an import
// path when it carries a version, or when its first element is a host name,
// which is what tells "github.com/titpetric/oida" from "cmd/go-fsck".
func isImportPath(source string) bool {
	if source == "" {
		return false
	}
	if info, err := os.Stat(source); err == nil && info.IsDir() {
		return false
	}
	if strings.Contains(source, "@") {
		return true
	}

	first, _, found := strings.Cut(source, "/")
	return found && strings.Contains(first, ".")
}

// resolveImportPath returns the directory a package of another module is in,
// downloading it into the module cache if it is not there yet.
//
// The resolution runs in a module of its own, in a temporary directory: the
// argument is a package path and "go mod download" takes a module path, so
// "github.com/titpetric/oida/model" is not something it can be handed. Asking
// go to require it and then asking where it went resolves both halves, the
// module that provides the package and the directory of the package itself.
//
// The version is what follows the at sign, and is "latest" without one.
func resolveImportPath(source string, verbose bool) (string, error) {
	path, version, found := strings.Cut(source, "@")
	if !found || version == "" {
		version = defaultVersion
	}

	work, err := os.MkdirTemp("", "go-fsck-extract-")
	if err != nil {
		return "", err
	}
	defer os.RemoveAll(work)

	run := func(args ...string) (string, error) {
		cmd := exec.Command("go", args...)
		cmd.Dir = work
		cmd.Env = append(os.Environ(), "GOFLAGS=-mod=mod")

		out, err := cmd.Output()
		if err != nil {
			message := strings.TrimSpace(string(err.(*exec.ExitError).Stderr))
			return "", fmt.Errorf("go %s: %s", strings.Join(args, " "), firstLine(message))
		}
		return strings.TrimSpace(string(out)), nil
	}

	if _, err := run("mod", "init", "go-fsck-extract"); err != nil {
		return "", err
	}
	if verbose {
		fmt.Printf("resolving %s@%s\n", path, version)
	}
	if _, err := run("get", path+"@"+version); err != nil {
		return "", err
	}

	dir, err := run("list", "-f", "{{.Dir}}", path)
	if err != nil {
		return "", err
	}
	if dir == "" {
		return "", fmt.Errorf("go list: %s@%s is in no directory", path, version)
	}

	// The module cache is read only, and every path under it is absolute.
	return filepath.Clean(dir), nil
}

// firstLine is the line of an error worth printing, since go writes several
// and the first names the failure.
func firstLine(message string) string {
	line, _, _ := strings.Cut(message, "\n")
	if line == "" {
		return message
	}
	return line
}
