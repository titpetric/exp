package main

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/titpetric/exp/cmd/go-ddd-stats/model"
)

func main() {
	if err := start(context.Background()); err != nil {
		log.Fatal(err)
	}
}

func start(_ context.Context) error {
	files, err := glob(".", ".go")
	if err != nil {
		return err
	}

	collection := &model.Stats{}
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

		collection.AppendFile(&model.File{
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

	collection.Histogram = model.Histogram(collection)

	enc := json.NewEncoder(os.Stdout)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", "  ")
	return enc.Encode(collection)
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
