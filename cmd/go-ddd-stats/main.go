package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/titpetric/exp/cmd/go-ddd-stats/model"
)

// options are the flags the command accepts. The renderers are exclusive: the
// last one named wins, and with none the whole collection is written as JSON.
type options struct {
	dir   string
	d2    bool
	stats bool
	json  bool
}

func main() {
	if err := start(context.Background()); err != nil {
		log.Fatal(err)
	}
}

func start(_ context.Context) error {
	opts, err := parseFlags(os.Args[1:])
	if err != nil {
		return err
	}

	stats, err := collect(opts.dir)
	if err != nil {
		return err
	}

	switch {
	case opts.d2:
		return renderD2(os.Stdout, stats)
	case opts.stats:
		return renderStats(os.Stdout, stats)
	default:
		return renderJSON(os.Stdout, stats)
	}
}

// parseFlags reads the command line. A lone "-" is accepted and ignored: it
// spells stdout, which is the only place output goes.
func parseFlags(args []string) (*options, error) {
	opts := &options{dir: "."}

	fs := flag.NewFlagSet("go-ddd-stats", flag.ContinueOnError)
	fs.BoolVar(&opts.d2, "d2", false, "write the size histogram as a d2 diagram")
	fs.BoolVar(&opts.stats, "stats", false, "write the size histogram and the counts, without the files")
	fs.BoolVar(&opts.json, "json", false, "write the whole collection as JSON (the default)")
	fs.Usage = func() {
		fmt.Fprintln(fs.Output(), "Usage: go-ddd-stats [flags] [path]")
		fmt.Fprintln(fs.Output())
		fmt.Fprintln(fs.Output(), "Report the size of every .go file below path, per file and per package,")
		fmt.Fprintln(fs.Output(), "with a histogram over them. The path defaults to the current directory.")
		fmt.Fprintln(fs.Output())
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return nil, err
	}

	for _, arg := range fs.Args() {
		if arg == "-" {
			continue
		}
		opts.dir = arg
	}
	return opts, nil
}

// collect walks the tree and measures every .go file in it. Vendored files are
// left out: they are somebody else's code and would swamp the histogram.
func collect(dir string) (*model.Stats, error) {
	files, err := glob(dir, ".go")
	if err != nil {
		return nil, err
	}

	collection := &model.Stats{}
	for _, filename := range files {
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
	return collection, nil
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
