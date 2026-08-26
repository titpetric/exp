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

// JSONSchema represents a JSON Schema document according to the draft-07 specification.
// It includes standard fields used to define types, formats, validations.
type JSONSchema struct {
	// Schema specifies the JSON Schema version URL.
	// Example: "http://json-schema.org/draft-07/schema#"
	Schema string `json:"$schema,omitempty"`
	// Ref is used to reference another schema definition.
	// Example: "#/definitions/SomeType"
	Ref string `json:"$ref,omitempty"`
	// Definitions contains subSchema definitions that can be referenced by $ref.
	Definitions map[string]*JSONSchema `json:"definitions,omitempty"`
	// Type indicates the JSON type of the instance (e.g., "object", "array", "string").
	Type string `json:"type,omitempty"`
	// Format provides additional semantic validation for the instance.
	// Common formats include "date-time", "email", etc.
	Format string `json:"format,omitempty"`
	// Pattern defines a regular expression that a string value must match
	Pattern string `json:"pattern,omitempty"`
	// Properties defines the fields of an object and their corresponding schemas
	Properties map[string]*JSONSchema `json:"properties,omitempty"`
	// Items defines the schema for array elements
	Items *JSONSchema `json:"items,omitempty"`
	// Enum restricts a value to a fixed set of values
	Enum []any `json:"enum,omitempty"`
	// Required lists the properties that must be present in an object
	Required []string `json:"required,omitempty"`
	// Description provides a human-readable explanation of the schema.
	Description string `json:"description,omitempty"`
	// Minimum specifies the minimum numeric value allowed.
	Minimum *float64 `json:"minimum,omitempty"`
	// Maximum specifies the maximum numeric value allowed.
	Maximum *float64 `json:"maximum,omitempty"`
	// ExclusiveMinimum, if true, requires the instance to be greater than (not equal to) Minimum.
	ExclusiveMinimum *bool `json:"exclusiveMinimum,omitempty"`
	// ExclusiveMaximum, if true, requires the instance to be less than (not equal to) Maximum.
	ExclusiveMaximum *bool `json:"exclusiveMaximum,omitempty"`
	// MultipleOf indicates that the numeric instance must be a multiple of this value.
	MultipleOf *float64 `json:"multipleOf,omitempty"`
	// AdditionalProperties controls whether an object can have properties beyond those defined
	// Can be a boolean or a schema that additional properties must conform to
	AdditionalProperties any `json:"additionalProperties,omitempty"`
}

// Module holds the go.mod of the module a package was extracted from.
//
// A package is only half of what a consumer takes on. The other half is the
// dependency set it drags in, and a release that moves a requirement to
// another major version, or replaces one with a fork, changes what it costs to
// depend on the package even though no symbol moved. Carrying the go.mod in
// the model is what lets two revisions be compared on that.
//
// The lists are sorted, so a go.mod that only had its blocks reordered
// produces the same model.
type Module struct {
	// Path is the module path, as the module directive declares it.
	Path string

	// GoVersion is the language version of the go directive. It bounds which
	// toolchains can build the module at all, so raising it drops support for
	// every consumer still on an older one.
	GoVersion string `json:",omitempty"`

	// Toolchain is the toolchain directive, and is empty for a module that
	// pins none.
	Toolchain string `json:",omitempty"`

	// Requires are the modules this one depends on.
	Requires []Require `json:",omitempty"`

	// Replaces are the replace directives.
	Replaces []Replace `json:",omitempty"`
}

// Replace is one replace directive of a go.mod.
//
// A replace changes what the build resolves a requirement to without changing
// what the requirement says, so it is recorded on its own: a release that
// starts or stops replacing a module changes what a consumer builds against
// while every version in the require block stays put.
type Replace struct {
	// Path is the module being replaced, and Version the single version the
	// directive applies to. An empty Version replaces every version of it.
	Path    string
	Version string `json:",omitempty"`

	// NewPath is what the module resolves to, which is either another module
	// path or a directory on disk. NewVersion is empty for a directory, which
	// carries no version.
	NewPath    string
	NewVersion string `json:",omitempty"`
}

// Require is one requirement of a go.mod.
type Require struct {
	// Path is the required module, and Version the version required of it.
	Path    string
	Version string

	// Indirect reports a requirement the module does not import itself, and
	// carries only to pin what a dependency of a dependency resolves to. It is
	// the "// indirect" comment go mod tidy writes. An indirect requirement is
	// bookkeeping, where a direct one is a decision.
	Indirect bool `json:",omitempty"`
}
