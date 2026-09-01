package files

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestFile_Flush covers what a restored file reads as: the package clause, the
// imports it was given, and the declarations, formatted the way gofmt writes
// them and grouped the way goimports groups them.
func TestFile_Flush(t *testing.T) {
	name := filepath.Join(t.TempDir(), "test_file.go")

	file := &File{
		Filename: name,
		Package:  "main",
		Imports:  []string{`"fmt"`, `"os"`, `"golang.org/x/crypto/bcrypt"`},
		Types: []string{
			"type MyType struct {\nName string\n}",
			"// AnotherType is documented.\ntype AnotherType struct {\nAge int\n}",
		},
	}

	if err := file.Flush(); err != nil {
		t.Fatalf("Flush() error = %v", err)
	}

	content, err := os.ReadFile(name)
	if err != nil {
		t.Fatalf("reading the file: %v", err)
	}

	want := `package main

import (
	"fmt"
	"os"

	"golang.org/x/crypto/bcrypt"
)

type MyType struct {
	Name string
}

// AnotherType is documented.
type AnotherType struct {
	Age int
}
`
	if string(content) != want {
		t.Errorf("file content mismatch.\nExpected:\n%s\nGot:\n%s", want, content)
	}
}

// TestFile_FlushDoc covers the package comment, which the model records as
// prose and a file carries as a comment.
func TestFile_FlushDoc(t *testing.T) {
	name := filepath.Join(t.TempDir(), "doc.go")

	file := &File{
		Filename: name,
		Package:  "model",
		Doc:      "Package model holds the recorded data.\n\nIt depends on nothing else.",
	}

	if err := file.Flush(); err != nil {
		t.Fatalf("Flush() error = %v", err)
	}

	content, err := os.ReadFile(name)
	if err != nil {
		t.Fatalf("reading the file: %v", err)
	}

	want := `// Package model holds the recorded data.
//
// It depends on nothing else.
package model
`
	if string(content) != want {
		t.Errorf("doc.go reads:\n%s\nwant:\n%s", content, want)
	}
}

// TestFile_FlushUnparseable covers the file the formatter cannot read: it is
// written as it was built, and the error names it.
func TestFile_FlushUnparseable(t *testing.T) {
	name := filepath.Join(t.TempDir(), "broken.go")

	file := &File{
		Filename: name,
		Package:  "main",
		Types:    []string{"func Open( {"},
	}

	err := file.Flush()
	if err == nil {
		t.Fatal("Flush() accepted source that does not parse")
	}
	if !strings.Contains(err.Error(), "broken.go") {
		t.Errorf("error = %v, want the file it could not format", err)
	}

	content, err := os.ReadFile(name)
	if err != nil {
		t.Fatalf("Flush() wrote nothing to look at: %v", err)
	}
	if !strings.Contains(string(content), "func Open( {") {
		t.Errorf("the file does not hold what was built:\n%s", content)
	}
}

// TestFile_FlushEmpty covers the file with nothing in it, which is not a file.
func TestFile_FlushEmpty(t *testing.T) {
	name := filepath.Join(t.TempDir(), "empty.go")

	if err := (&File{Filename: name, Package: "main"}).Flush(); err != nil {
		t.Fatalf("Flush() error = %v", err)
	}
	if _, err := os.Stat(name); err == nil {
		t.Error("Flush() wrote a file holding nothing")
	}
}
