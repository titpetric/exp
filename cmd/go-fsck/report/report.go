package report

import (
	"fmt"

	"github.com/titpetric/exp/cmd/go-fsck/internal"
	"github.com/titpetric/exp/cmd/go-fsck/report/testusage"
	"github.com/titpetric/tools/splint/model"
)

func getDefinitions(cfg *options) (model.DefinitionList, error) {
	return internal.DefinitionsFrom(cfg.inputFile, internal.Options(".", false, false, false, cfg.verbose))
}

func report(cfg *options) error {
	defs, err := getDefinitions(cfg)
	if len(defs) == 0 || err != nil {
		return fmt.Errorf("error getting definitions: %w, len %d", err, len(defs))
	}

	report, err := testusage.NewReport(defs)
	if err != nil {
		return fmt.Errorf("error generating report: %w", err)
	}

	fmt.Println(report)
	return nil
	// return json.NewEncoder(os.Stdout).Encode(report)
}
