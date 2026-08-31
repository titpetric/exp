package query

import (
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/fbiville/markdown-table-formatter/pkg/markdown"

	"github.com/titpetric/exp/cmd/go-fsck/internal"
	"github.com/titpetric/tools/splint/model"
)

func getDefinitions(cfg *options) (model.DefinitionList, error) {
	return internal.DefinitionsFrom(cfg.inputFile, internal.Options(".", true, false, false, cfg.verbose))
}

func query(cfg *options) error {
	defs, err := getDefinitions(cfg)
	if err != nil {
		return err
	}

	// Loop through function definitions and collect referenced
	// symbols from imported packages. Globals may also reference
	// imported packages so this is incomplete at the moment.

	results := model.DeclarationList{}

	functionSignature := []string{"http.ResponseWriter", "*http.Request"}

	var middleware = struct {
		Arguments []string
		Returns   []string
	}{
		Arguments: []string{"http.ResponseWriter", "*http.Request", ""},
		Returns:   []string{"error", "int"},
	}

	for _, def := range defs {
		for _, fn := range def.Funcs {
			if cfg.showHandlers {
				if !reflect.DeepEqual(fn.Arguments, functionSignature) {
					continue
				}
			}

			if cfg.showMiddleware {
				if !reflect.DeepEqual(fn.Arguments, middleware.Arguments) {
					continue
				}
				if !reflect.DeepEqual(fn.Returns, middleware.Returns) {
					continue
				}
			}

			results = append(results, fn)
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
		table = append(table, []string{result.Name, result.File, result.Receiver, result.Signature})
	}

	t, err := markdown.NewTableFormatterBuilder().WithPrettyPrint().Build("Function", "File", "Receiver", "Signature").Format(table)
	if err != nil {
		return err
	}

	fmt.Println(t)

	return nil
}
