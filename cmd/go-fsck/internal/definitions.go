package internal

import (
	"context"
	"os"
	"os/signal"

	"github.com/titpetric/tools/splint"
	"github.com/titpetric/tools/splint/analyzer"
	"github.com/titpetric/tools/splint/loader"
	"github.com/titpetric/tools/splint/model"
)

// Definitions reads a tree into the packages it declares.
//
// Every subcommand of go-fsck needs this and each used to carry its own copy
// of it, which is twenty lines repeated nine times and nine places for the
// listing to drift. There is one now, and it is a call into splint: the
// parsing left this repository and what is left here is the reading of it.
func Definitions(options splint.Options) (model.DefinitionList, error) {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	doc, err := analyzer.New(options).Parse(ctx)
	if err != nil {
		return nil, err
	}
	return doc.Packages, nil
}

// DefinitionsFrom reads a tree, or reads back a document already written for
// it. A subcommand that takes an input file is answering the same question off
// a file rather than off the source, and a file that will not open means the
// source is what it has to read.
func DefinitionsFrom(filename string, options splint.Options) (model.DefinitionList, error) {
	if filename != "" {
		if doc, err := loader.Load(filename); err == nil {
			return doc.Packages, nil
		}
	}
	return Definitions(options)
}

// Options is what a parse of one tree is asked for, built from the flags a
// subcommand carries.
func Options(sourcePath string, recursive, includeTests, includeSources, verbose bool) splint.Options {
	pattern := "."
	if recursive {
		pattern = "./..."
	}
	if sourcePath == "" {
		sourcePath = "."
	}

	return splint.Options{
		SourcePath:     sourcePath,
		Pattern:        pattern,
		IncludeTests:   includeTests,
		IncludeSources: includeSources,
		Verbose:        verbose,
	}
}
