package coverfunc

import (
	"fmt"
)

// CoverageInfo represents information about coverage for a specific function.
type CoverageInfo struct {
	File      string `json:",omitempty"`
	Filename  string `json:",omitempty"`
	Package   string
	Line      int    `json:",omitempty"`
	Function  string `json:",omitempty"`
	Functions int    `json:",omitempty"`
	Coverage  float64
}

// FunctionInfo holds coverage info for functions.
type FunctionInfo CoverageInfo
type PackageInfo CoverageInfo
type FileInfo CoverageInfo

// String returns a string representation of a FunctionInfo.
func (p FunctionInfo) String() string {
	return fmt.Sprintf("%s, file %s, function %s, coverage %.2f%%", p.Package, p.Filename, p.Function, p.Coverage)
}

// String returns a string representation of a PackageInfo.
func (p PackageInfo) String() string {
	return fmt.Sprintf("%s, symbols %d, coverage %.2f%%", p.Package, p.Functions, p.Coverage)
}

// String returns a string representation of a FileInfo.
func (f FileInfo) String() string {
	return fmt.Sprintf("%s, symbols %d, coverage %.2f%%", f.Filename, f.Functions, f.Coverage)
}
