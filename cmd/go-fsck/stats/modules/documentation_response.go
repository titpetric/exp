package modules

import (
	"fmt"
)

// DocumentationResponse holds the metrics associated with a DefinitionList.
type DocumentationResponse struct {
	Packages    int
	PackageDocs int

	Symbols      int
	SymbolDocs   int
	SymbolGodocs int
}

// NewDocumentationResponse will return a zero value of DocumentationResponse{}.
func NewDocumentationResponse() DocumentationResponse {
	return DocumentationResponse{}
}

// String prints a simple representation of the documentation response.
func (r DocumentationResponse) String() string {
	return fmt.Sprintf(
		"Of %d symbols, %d have comments, and %d follow godoc standards. Out of %d packages, %d have package docs.",
		r.Symbols, r.SymbolDocs, r.SymbolGodocs, r.Packages, r.PackageDocs,
	)
}

// Merge will take a definition level documentation response and merge it to the receiver.
func (r *DocumentationResponse) Merge(in DocumentationResponse) {
	r.Packages += in.Packages
	r.PackageDocs += in.PackageDocs
	r.Symbols += in.Symbols
	r.SymbolDocs += in.SymbolDocs
	r.SymbolGodocs += in.SymbolGodocs
}
