package model

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
