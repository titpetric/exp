package model

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
