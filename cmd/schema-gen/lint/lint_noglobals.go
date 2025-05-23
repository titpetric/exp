package lint

import "github.com/titpetric/exp/cmd/schema-gen/model"

func linterNoGlobals(cfg *options, pkgInfo *model.PackageInfo) *LintError {
	// require-no-globals
	//
	// When doing model/repository work, the package `model` provides
	// the type declarations for `structs`, `var` and `const`'s.
	//
	// The repository package should provide interfaces, possibly
	// constructor functions for the repositories, but should not
	// expose symbols outside of that. Those belong in model.

	return nil
}
