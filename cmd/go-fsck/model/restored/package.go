package model

import (
	"fmt"

	"golang.org/x/tools/go/packages"
)

// Package holds go package information.
type Package struct {
	// ID is the ID of the package as x/tools packages loads it.
	ID string
	// Package is the name of the package.
	Package string
	// ImportPath contains the import path (github...).
	ImportPath string
	// Path is sanitized to contain the relative location (folder).
	Path string
	// TestPackage is true if this is a test package.
	TestPackage bool

	// Complexity to collect test coverage on package.
	Complexity *Complexity `json:",omitempty"`

	// Pkg serves to carry ast package information, preventing a double Load().
	// It's used during analysis and merging and discarded for the result.
	Pkg *packages.Package `json:"-"`
}

func (p Package) Equal(in Package) bool {
	return p.ImportPath == in.ImportPath
}

func (p Package) Name() string {
	return p.Package
}

func (p Package) String() string {
	return fmt.Sprintf("package=%s import_path=%s path=%s test_package=%v", p.Package, p.ImportPath, p.Path, p.TestPackage)
}
