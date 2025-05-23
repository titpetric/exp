package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
)

func main() {
	if err := start(context.Background()); err != nil {
		log.Fatal(err)
	}
}

type ModuleStats struct {
	Files    []*File
	Packages []*Package
}

func (m *ModuleStats) AppendFile(f *File) {
	m.Files = append(m.Files, f)
}

func (m *ModuleStats) Package(packagePath string) *Package {
	for _, pkg := range m.Packages {
		if pkg.Path == packagePath {
			return pkg
		}
	}
	result := &Package{
		Path: packagePath,
	}
	m.Packages = append(m.Packages, result)
	return result
}

type Package struct {
	Name    string
	Path    string
	Size    int64
	Count   int64
	Average int64
}

type File struct {
	Name    string
	Path    string
	Package string
	Size    int64
}

type Size struct {
	Size  string
	Count int64
}

func start(_ context.Context) error {
	files, err := glob(".", ".go")
	if err != nil {
		return err
	}

	collection := ModuleStats{}
	for _, filename := range files {
		// skip vendored files from analysis
		if strings.Contains(filename, "vendor/") {
			continue
		}

		size, _ := filesize(filename)

		packagePath := path.Dir(filename)
		if packagePath == "." {
			packagePath = ""
		}

		collection.AppendFile(&File{
			Name:    filename,
			Path:    packagePath,
			Package: path.Base(path.Dir(filename)),
			Size:    size,
		})
	}

	for _, record := range collection.Files {
		pkg := collection.Package(record.Path)
		pkg.Name = record.Package
		pkg.Size += record.Size
		pkg.Count++
		pkg.Average = pkg.Size / pkg.Count
	}

	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(collection)
}

func getPackageName() string {
	output, _ := exec.Command("go", "list", ".").CombinedOutput()
	return strings.TrimSpace(string(output))
}

func getPackageCount() int {
	output, _ := exec.Command("go", "list", "./...").CombinedOutput()
	lines := bytes.Split(output, []byte("\n"))
	return len(lines)
}

func findGroup(file File) string {
	var increment int64 = 4

	bucket := increment
	for {
		if file.Size > bucket*1024 {
			bucket *= 2
			continue
		}
		return fmt.Sprint(bucket)
	}
}

func filesize(filename string) (int64, error) {
	fi, err := os.Stat(filename)
	if err != nil {
		return 0, err
	}
	return fi.Size(), nil
}

func glob(dir string, ext string) ([]string, error) {
	files := []string{}
	err := filepath.Walk(dir, func(filename string, f os.FileInfo, err error) error {
		if filepath.Ext(filename) == ext {
			files = append(files, filename)
		}
		return nil
	})

	return files, err
}
