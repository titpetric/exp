package search

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/fbiville/markdown-table-formatter/pkg/markdown"

	"github.com/titpetric/exp/cmd/go-fsck/internal"
	"github.com/titpetric/tools/splint/model"
)

func getDefinitions(cfg *options) (model.DefinitionList, error) {
	return internal.DefinitionsFrom(cfg.inputFile, internal.Options(".", false, true, false, cfg.verbose))
}

func search(cfg *options) error {
	defs, err := getDefinitions(cfg)
	if err != nil {
		return err
	}

	ref := strings.Split(cfg.reference, ".")[0]

	// Loop through function definitions and collect referenced
	// symbols from imported packages. Globals may also reference
	// imported packages so this is incomplete at the moment.

	results := model.DeclarationList{}

	for _, def := range defs {
		for _, fn := range def.Funcs {
			if cfg.name != "" && !strings.Contains(fn.Name, cfg.name) {
				continue
			}

			if ref == "" {
				results = append(results, fn)
				continue
			}

			symbols, ok := fn.References[ref]
			if ok {
				// clear other references for results
				fn.References = map[string][]string{
					ref: symbols,
				}
				results = append(results, fn)
			}
		}
	}

	// Encode aggregated results as json.
	if cfg.json {
		b, err := json.Marshal(results)
		if err != nil {
			return err
		}
		fmt.Println(string(b))
		return nil
	}

	// Encode aggregated results as markdown.
	table := [][]string{}
	for _, result := range results {
		table = append(table, []string{result.Name, result.File, strings.Join(result.References[ref], ", ")})
	}

	t, err := markdown.NewTableFormatterBuilder().WithPrettyPrint().Build("Function", "File", "Symbols").Format(table)
	if err != nil {
		return err
	}

	fmt.Println(t)

	return nil
}
