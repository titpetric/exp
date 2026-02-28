package lint

import (
	"encoding/json"
	"errors"
	"fmt"

	"gopkg.in/yaml.v3"

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
			totalSymbols := len(defs)
			if len(importsIssues) > 0 {
				hasErrors = true
				if cfg.jsonOut {
					allIssues = append(allIssues, map[string]interface{}{
						"rule":   "import-collision",
						"issues": importsIssues,
					})
				} else if cfg.summarize {
					stats := importsLinter.GetStatistics(totalSymbols)
					yamlData, _ := yaml.Marshal(map[string]interface{}{
						"imports": stats,
					})
					fmt.Print(string(yamlData))
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
			totalSymbols := len(defs) // Count all definitions as symbols for godoc
			if len(issues) > 0 {
				hasErrors = true
				if cfg.jsonOut {
					allIssues = append(allIssues, map[string]interface{}{
						"rule":   "godoc",
						"issues": issues,
					})
				} else if cfg.summarize {
					stats := linter.GetStatistics(totalSymbols)
					yamlData, _ := yaml.Marshal(map[string]interface{}{
						"godoc": stats,
					})
					fmt.Print(string(yamlData))
				} else {
					for _, issue := range issues {
						fmt.Println(issue.String())
					}
				}
			}

		case "func-args":
			linter := rules.NewFuncArgsLinter()
			linter.Lint(defs)
			issues := linter.Issues()
			if len(issues) > 0 {
				hasErrors = true
				if cfg.jsonOut {
					allIssues = append(allIssues, map[string]interface{}{
						"rule":    "func-args",
						"issues":  issues,
						"summary": linter.IssueSummary(),
					})
				} else if !cfg.summarize {
					for _, issue := range issues {
						fmt.Println(issue.String())
					}
				}
			}
			if cfg.summarize {
				stats := linter.GetStatistics(len(defs))
				yamlData, _ := yaml.Marshal(map[string]interface{}{
					"func-args": stats,
				})
				fmt.Print(string(yamlData))
			}

		case "func-returns":
			linter := rules.NewFuncReturnsLinter()
			linter.Lint(defs)
			issues := linter.Issues()
			if len(issues) > 0 {
				hasErrors = true
				if cfg.jsonOut {
					allIssues = append(allIssues, map[string]interface{}{
						"rule":   "func-returns",
						"issues": issues,
					})
				} else if !cfg.summarize {
					for _, issue := range issues {
						fmt.Println(issue.String())
					}
				}
			}
			if cfg.summarize {
				stats := linter.GetStatistics(len(defs))
				yamlData, _ := yaml.Marshal(map[string]interface{}{
					"func-returns": stats,
				})
				fmt.Print(string(yamlData))
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
