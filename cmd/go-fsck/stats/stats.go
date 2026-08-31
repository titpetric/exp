package stats

import (
	"encoding/json"
	"fmt"

	"github.com/titpetric/exp/cmd/go-fsck/internal"
	"github.com/titpetric/exp/cmd/go-fsck/stats/modules"
	"github.com/titpetric/tools/splint/model"
)

func getDefinitions(cfg *options) (model.DefinitionList, error) {
	return internal.DefinitionsFrom(cfg.inputFile, internal.Options(".", false, true, false, cfg.verbose))
}

func report(title string, value any) {
	fmt.Println("##", title)
	fmt.Println()
	if v, ok := value.(fmt.Stringer); ok {
		fmt.Println(v.String())
	} else {
		j, _ := json.MarshalIndent(value, "", "  ")
		fmt.Println(string(j))
	}
	fmt.Println()
}

func stats(cfg *options) error {
	defs, err := getDefinitions(cfg)
	if err != nil {
		return err
	}

	report("Documentation", modules.Documentation(defs))
	report("Package stats", modules.PackageStats(defs))
	report("Import usage", modules.ImportStats(defs))
	report("Reverse symbol usage", modules.ReverseUsage(defs))

	return nil
}
