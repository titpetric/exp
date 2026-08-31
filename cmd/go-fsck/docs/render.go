package docs

import (
	"encoding/json"
	"os"

	"github.com/titpetric/exp/cmd/go-fsck/internal"
	"github.com/titpetric/tools/splint/model"
)

func getDefinitions(cfg *options) (model.DefinitionList, error) {
	return internal.DefinitionsFrom(cfg.inputFile, internal.Options(".", false, false, false, cfg.verbose))
}

func render(cfg *options) error {
	defs, err := getDefinitions(cfg)
	if err != nil {
		return err
	}

	if cfg.split {
		return renderSplit(cfg, defs)
	}

	switch cfg.render {
	case "spec":
		return renderSpec(cfg, defs)
	case "imports":
		return renderImports(cfg, defs)
	case "json":
		return renderJSON(cfg, defs)
	case "puml", "plantuml":
		return renderPlantUML(cfg, defs)
	default:
		return renderMarkdown(cfg, defs)
	}
}

func renderJSON(_ *options, defs []*model.Definition) error {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(defs)
}
