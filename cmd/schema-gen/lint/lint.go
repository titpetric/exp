package lint

import (
	"fmt"
	"os"
	"strings"

	"github.com/titpetric/exp/cmd/schema-gen/model"
	. "github.com/titpetric/exp/cmd/schema-gen/model"
)

func lint(cfg *options) error {
	pkgInfos, err := model.Load(cfg.inputFile)
	if err != nil {
		return fmt.Errorf("Error loading package info: %w", err)
	}

	for _, pkgInfo := range pkgInfos {

		errs := NewLintError()
		errs.Combine(runLinter(cfg, NewLinter("lint structs", linterStructs), pkgInfo))
		errs.Combine(runLinter(cfg, NewLinter("lint fields", linterFields), pkgInfo))
		errs.Combine(runLinter(cfg, NewLinter("lint globals", linterNoGlobals), pkgInfo))

		if errs.Empty() {
			return nil
		}

		fmt.Println(errs.Error())

		if cfg.summary {
			rules := map[string]int{}
			for _, err := range errs.errs {
				errStr := strings.SplitN(err, "\n", 2)
				errFields := strings.Fields(errStr[0])
				rule := strings.Join(errFields[1:], " ")
				rules[rule] = rules[rule] + 1
			}

			for rule, count := range rules {
				fmt.Printf("- %d %s\n", count, rule)
			}
		}

		os.Exit(1)
	}

	return nil
}

func runLinter(cfg *options, linter Linter, pkgInfo *PackageInfo) *LintError {
	return linter.Do(cfg, pkgInfo)
}
