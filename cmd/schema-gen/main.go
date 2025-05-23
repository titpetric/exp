package main

// This app parses structs from a package in order to extract
// known data model struct fields into a json document.

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"golang.org/x/exp/maps"

	"github.com/titpetric/exp/cmd/schema-gen/extract"
	"github.com/titpetric/exp/cmd/schema-gen/jsonschema"
	"github.com/titpetric/exp/cmd/schema-gen/lint"
	"github.com/titpetric/exp/cmd/schema-gen/list"
	"github.com/titpetric/exp/cmd/schema-gen/markdown"
	"github.com/titpetric/exp/cmd/schema-gen/restore"
)

func main() {
	if err := start(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func start() (err error) {
	commands := map[string]func() error{
		"extract":    extract.Run,
		"restore":    restore.Run,
		"markdown":   markdown.Run,
		"lint":       lint.Run,
		"list":       list.Run,
		"jsonschema": jsonschema.Run,
	}
	commandList := maps.Keys(commands)
	sort.Strings(commandList)

	if len(os.Args) < 2 {
		fmt.Println("Usage: schema-gen <command> help")
		fmt.Printf("Available commands: %s\n", strings.Join(commandList, ", "))
		return nil
	}

	commandFn, ok := commands[os.Args[1]]
	if ok {
		return commandFn()
	}

	return fmt.Errorf("Unknown command: %q", os.Args[1])
}
