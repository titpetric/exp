package model

type Complexity struct {
	Cognitive  int
	Cyclomatic int
	Lines      int
}

// FieldList contains all struct fields.
type FieldList []*Field
