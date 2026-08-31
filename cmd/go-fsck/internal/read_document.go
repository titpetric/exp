package internal

import (
	"github.com/titpetric/tools/splint/loader"
	"github.com/titpetric/tools/splint/model"
)

// ReadDocument reads a document back from a file, and returns the packages in
// it. The document a parse wrote is what a comparison and a restore read: both
// work off a model rather than off source.
func ReadDocument(filename string) (model.DefinitionList, error) {
	doc, err := loader.Load(filename)
	if err != nil {
		return nil, err
	}
	return doc.Packages, nil
}
