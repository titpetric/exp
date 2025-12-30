package lint

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/titpetric/exp/cmd/go-fsck/internal"
	"github.com/titpetric/exp/cmd/go-fsck/lint/rules"
	"github.com/titpetric/exp/cmd/go-fsck/model"
	"github.com/titpetric/exp/cmd/go-fsck/model/loader"
)

func getDefinitions(cfg *options) ([]*model.Definition, error) {
	// list current local packages
	pattern := "./..."
	if len(cfg.args) > 0 {
		pattern = cfg.args[0]
	}

	packages, err := internal.ListPackages(".", pattern)
	if err != nil {
		return nil, err
	}

	defs := []*model.Definition{}
	getDef := func(in *model.Definition) *model.Definition {
		for _, def := range defs {
			if def.Package.Equal(in.Package) {
				return def
			}
		}
		return nil
	}

	for _, pkg := range packages {
		d, err := loader.Load(pkg, false, cfg.verbose)
		if err != nil {
			return nil, err
		}

		for _, in := range d {
			def := getDef(in)
			if def != nil {
				def.Merge(in)
				continue
			}
			defs = append(defs, in)
		}
	}

	return defs, nil
}

func lint(cfg *options) error {
	defs, err := getDefinitions(cfg)
	if err != nil {
		return err
	}

	var allIssues []interface{}
	hasErrors := false

	// Run enabled linters
	activeRules := cfg.GetRules()
	for _, ruleName := range activeRules {
		switch ruleName {
		case "imports":
			// Check import collisions (always enabled)
			importsLinter := rules.NewImportsLinter()
			importsLinter.Lint(defs)
			importsIssues := importsLinter.Issues()
			if len(importsIssues) > 0 {
				hasErrors = true
				if cfg.jsonOut {
					allIssues = append(allIssues, map[string]interface{}{
						"rule":   "import-collision",
						"issues": importsIssues,
					})
				} else {
					for _, err := range importsIssues {
						fmt.Println(err)
					}
				}
			}

		case "godoc":
			linter := rules.NewGodocLinter()
			linter.Lint(defs)
			issues := linter.Issues()
			if len(issues) > 0 {
				hasErrors = true
				if cfg.jsonOut {
					allIssues = append(allIssues, map[string]interface{}{
						"rule":   "godoc",
						"issues": issues,
					})
				} else if cfg.summarize {
					summary := linter.IssueSummary()
					fmt.Printf("Godoc linter summary:\n")
					for issueType, count := range summary {
						fmt.Printf("  - %d %s\n", count, issueType)
					}
				} else {
					for _, issue := range issues {
						fmt.Println(issue.String())
					}
				}
			}
		}
	}

	if cfg.jsonOut && len(allIssues) > 0 {
		jsonBytes, _ := json.MarshalIndent(allIssues, "", "  ")
		fmt.Println(string(jsonBytes))
	}

	if !hasErrors {
		return nil
	}

	return errors.New("Linter not passing")
}
