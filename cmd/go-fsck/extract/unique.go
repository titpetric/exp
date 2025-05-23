package extract

import "github.com/titpetric/exp/cmd/go-fsck/model"

func unique(defs []*model.Definition) []*model.Definition {
	result := make([]*model.Definition, 0, len(defs))

	for _, def := range defs {
		var match bool
		for _, res := range result {
			if res.ID == def.ID {
				match = true
				res.Merge(def)
				break
			}
		}

		if !match {
			result = append(result, def)
		}
	}

	return result
}
