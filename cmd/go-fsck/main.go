package main

import (
	"fmt"
	"os"
	"sort"
	"strings"

	_ "modernc.org/sqlite"

	"golang.org/x/exp/maps"

	"github.com/titpetric/exp/cmd/go-fsck/coverage"
	"github.com/titpetric/exp/cmd/go-fsck/docs"
	"github.com/titpetric/exp/cmd/go-fsck/edges"
	"github.com/titpetric/exp/cmd/go-fsck/extract"
	"github.com/titpetric/exp/cmd/go-fsck/lint"
	"github.com/titpetric/exp/cmd/go-fsck/query"
	"github.com/titpetric/exp/cmd/go-fsck/report"
	"github.com/titpetric/exp/cmd/go-fsck/restore"
	"github.com/titpetric/exp/cmd/go-fsck/search"
	"github.com/titpetric/exp/cmd/go-fsck/sqlite"
	"github.com/titpetric/exp/cmd/go-fsck/stats"
	"github.com/titpetric/exp/cmd/go-fsck/test"
)

func main() {
	if err := start(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func start() (err error) {
	commands := map[string]func() error{
		"extract":  extract.Run,
		"coverage": coverage.Run,
		"restore":  restore.Run,
		"stats":    stats.Run,
		"lint":     lint.Run,
		"search":   search.Run,
		"query":    query.Run,
		"docs":     docs.Run,
		"report":   report.Run,
		"sqlite":   sqlite.Run,
		"test":     test.Run,
		"edges":    edges.Run,
	}
	commandList := maps.Keys(commands)
	sort.Strings(commandList)

	if len(os.Args) < 2 {
		fmt.Println("Usage: go-fsck <command> help")
		fmt.Printf("Available commands: %s\n", strings.Join(commandList, ", "))
		return nil
	}

	commandFn, ok := commands[os.Args[1]]
	if ok {
		return commandFn()
	}

	return fmt.Errorf("Unknown command: %q", os.Args[1])
}
