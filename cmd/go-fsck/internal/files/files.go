package files

import (
	"fmt"
	"os"
	"strings"

	"golang.org/x/tools/imports"
)

// File struct represents a Go source file with its components.
type File struct {
	Filename string
	Package  string
	Imports  []string
	Types    []string

	// Doc is the package comment, and is written above the package clause of
	// the one file that carries it. The model records the text of a comment
	// without the markers, so they go back on here.
	Doc string
}

// Body is the file without its imports, which is what a reader of the syntax
// is given to find out which imports it needs.
func (f *File) Body() string {
	lines := make([]string, 0, len(f.Types)+2)

	if f.Doc != "" {
		lines = append(lines, comment(f.Doc))
	}
	lines = append(lines, "package "+f.Package, "")
	lines = append(lines, f.Types...)

	return strings.Join(lines, "\n") + "\n"
}

// Flush writes the file, formatted the way goimports would.
//
// The formatter is asked to format and not to fix: the imports written here
// are the ones the declarations reach, and a formatter resolving what it
// thinks is missing would reach for the network to do it. What it does do is
// group them, so the standard library sits apart from everything else.
//
// A file the formatter cannot parse is written as it was built and reported:
// the source is what a reader needs to see why, and an empty file says
// nothing.
func (f *File) Flush() error {
	if len(f.Types) == 0 && f.Doc == "" {
		return nil
	}

	var out strings.Builder

	if f.Doc != "" {
		out.WriteString(comment(f.Doc) + "\n")
	}
	out.WriteString("package " + f.Package + "\n\n")

	if len(f.Imports) > 0 {
		out.WriteString("import (\n")
		for _, one := range f.Imports {
			out.WriteString("\t" + one + "\n")
		}
		out.WriteString(")\n\n")
	}

	for _, body := range f.Types {
		out.WriteString(body + "\n\n")
	}

	source := []byte(out.String())
	formatted, err := imports.Process(f.Filename, source, &imports.Options{
		Comments:   true,
		TabIndent:  true,
		TabWidth:   8,
		FormatOnly: true,
	})
	if err != nil {
		if writeErr := os.WriteFile(f.Filename, source, 0o644); writeErr != nil {
			return writeErr
		}
		return fmt.Errorf("%s does not parse: %w", f.Filename, err)
	}

	return os.WriteFile(f.Filename, formatted, 0o644)
}

// comment writes a doc comment back the way it was read.
//
// The model records what a comment says and not the markers it says it
// through, so a package comment comes back as prose and goes out as a comment
// again. One that kept its markers is left as it is.
func comment(doc string) string {
	doc = strings.TrimSpace(doc)
	if strings.HasPrefix(doc, "//") || strings.HasPrefix(doc, "/*") {
		return doc
	}

	lines := strings.Split(doc, "\n")
	for i, line := range lines {
		if line = strings.TrimRight(line, " \t"); line == "" {
			lines[i] = "//"
			continue
		}
		lines[i] = "// " + line
	}
	return strings.Join(lines, "\n")
}
