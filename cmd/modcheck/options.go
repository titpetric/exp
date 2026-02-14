package main

import (
	"fmt"

	flag "github.com/spf13/pflag"
)

type options struct {
	suggest    bool
	forUpgrade bool
	json       bool
	skip       []string
	goModPath  string
	args       []string
}

func NewOptions() *options {
	cfg := &options{
		goModPath: "go.mod",
	}

	flag.BoolVar(&cfg.suggest, "suggest", cfg.suggest, "print go get commands to update dependencies")
	flag.BoolVar(&cfg.forUpgrade, "for-upgrade", cfg.forUpgrade, "only list packages for upgrade")
	flag.BoolVar(&cfg.json, "json", cfg.json, "output as JSON")
	flag.StringSliceVar(&cfg.skip, "skip", cfg.skip, "skip packages")
	flag.Parse()

	cfg.args = flag.Args()

	return cfg
}

func PrintHelp() {
	fmt.Println("Usage: schema-gen restore <options>:")
	fmt.Println()
	flag.PrintDefaults()
}
