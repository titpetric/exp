package model

type Complexity struct {
	Cognitive  int
	Cyclomatic int
	Lines      int

	// Coverage is filled out of band (summary coverfunc).
	Coverage float64 `json:",omitempty"`
}

// FieldList contains all struct fields.
type FieldList []*Field
