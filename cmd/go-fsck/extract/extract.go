package extract

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/titpetric/exp/cmd/go-fsck/internal"
	"github.com/titpetric/tools/splint/model"
)

// getDefinitions reads the tree the options point at.
//
// Everything this used to do, walking the module tree and merging what came
// back, is what splint's analyzer does now. What is left here is the command:
// the flags, and writing the result out.
func getDefinitions(cfg *options) (model.DefinitionList, error) {
	return internal.Definitions(internal.Options(cfg.sourcePath, cfg.recursive, cfg.includeTests, cfg.includeSources, cfg.verbose))
}

func extract(cfg *options) error {
	// A source path that names a package of another module is read out of the
	// module cache, which is what makes an upstream package extractable
	// without checking it out first.
	if isImportPath(cfg.sourcePath) {
		dir, err := resolveImportPath(cfg.sourcePath, cfg.verbose)
		if err != nil {
			return err
		}
		cfg.sourcePath = dir
	}

	definitions, err := getDefinitions(cfg)
	if err != nil {
		return err
	}

	output := os.Stdout
	switch cfg.outputFile {
	case "", "-":
	default:
		fmt.Println(cfg.outputFile)
		var err error
		output, err = os.Create(cfg.outputFile)
		if err != nil {
			return err
		}
		defer output.Close()
	}

	encoder := json.NewEncoder(output)
	if cfg.prettyJSON {
		encoder.SetIndent("", "  ")
	}

	return encoder.Encode(definitions)
}
