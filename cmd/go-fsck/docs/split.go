package docs

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/titpetric/exp/cmd/go-fsck/model"
)

// PackageDoc holds a single package's documentation and metadata.
type PackageDoc struct {
	ImportPath string
	Filename   string
	Content    string
}

// generateFilename creates a filename from an import path.
// If strip is provided, it removes that prefix from the import path,
// otherwise splits by "/" and excludes the first element.
func generateFilename(importPath, strip string) string {
	path := importPath

	if strip != "" {
		if strings.HasPrefix(importPath, strip) {
			path = strings.TrimPrefix(importPath, strip)
			path = strings.TrimPrefix(path, "/")
		}
	} else {
		// Split by "/" and exclude the first element
		parts := strings.Split(importPath, "/")
		if len(parts) > 1 {
			path = strings.Join(parts[1:], "/")
		}
	}

	// Handle empty path (root package)
	if path == "" {
		path = filepath.Base(importPath)
	}

	// Replace "/" with "_" and add .md extension
	filename := strings.ReplaceAll(path, "/", "_") + ".md"
	return filename
}

// groupDefinitionsByPackage organizes definitions by their import path.
func groupDefinitionsByPackage(defs []*model.Definition) map[string][]*model.Definition {
	groups := make(map[string][]*model.Definition)
	for _, def := range defs {
		// Skip test packages
		if def.Package.TestPackage || strings.HasSuffix(def.Package.Package, "_test") {
			continue
		}
		groups[def.Package.ImportPath] = append(groups[def.Package.ImportPath], def)
	}
	return groups
}

// renderMarkdownForPackage renders markdown for a single package.
func renderMarkdownForPackage(defs []*model.Definition) string {
	if len(defs) == 0 {
		return ""
	}

	var buf strings.Builder

	def := defs[0] // All definitions in this group are from the same package

	var (
		types  = def.Types.Exported()
		consts = def.Consts.Exported()
		vars   = def.Vars.Exported()
		funcs  = def.Funcs.Exported()
	)

	var packageName = def.Package.Path
	if packageName == "." {
		packageName = filepath.Base(def.Package.ImportPath)
	}

	fmt.Fprintf(&buf, "# Package %s\n\n", packageName)
	fmt.Fprintf(&buf, "```go\n")
	fmt.Fprintf(&buf, "import (\n\t\"%s\"\n}\n", def.Package.ImportPath)
	fmt.Fprintf(&buf, "```\n\n")

	if def.Doc != "" {
		fmt.Fprintf(&buf, "%s\n\n", strings.TrimSpace(def.Doc))
	}

	if len(types) > 0 {
		fmt.Fprint(&buf, "## Types\n\n")
		for _, v := range types {
			src := strings.TrimSpace(v.Source)
			fmt.Fprintf(&buf, "```go\n%s\n```\n\n", src)
		}
	}

	if len(consts) > 0 {
		fmt.Fprint(&buf, "## Consts\n\n")
		for _, v := range consts {
			src := strings.TrimSpace(v.Source)
			fmt.Fprintf(&buf, "```go\n%s\n```\n\n", src)
		}
	}

	if len(vars) > 0 {
		fmt.Fprint(&buf, "## Vars\n\n")
		for _, v := range vars {
			src := strings.TrimSpace(v.Source)
			fmt.Fprintf(&buf, "```go\n%s\n```\n\n", src)
		}
	}

	symbol := func(fn *model.Declaration) string {
		if fn.Receiver != "" {
			return "func (" + fn.Receiver + ") " + fn.Signature
		}
		return "func " + fn.Signature
	}

	if len(funcs) > 0 {
		fmt.Fprint(&buf, "## Function symbols\n\n")

		for _, fn := range funcs {
			fmt.Fprintf(&buf, "- `%s`\n", symbol(fn))
		}
		fmt.Fprint(&buf, "\n")

		// Documented functions first.
		for _, fn := range funcs {
			if fn.Doc == "" {
				continue
			}

			fmt.Fprintf(&buf, "### %s\n\n", fn.Name)
			fmt.Fprintf(&buf, "%s\n\n", strings.TrimSpace(fn.Doc))
			fmt.Fprintf(&buf, "```go\n%s\n```\n\n", symbol(fn))
		}

		// List undocumented ones.
		for _, fn := range funcs {
			if fn.Doc != "" {
				continue
			}

			fmt.Fprintf(&buf, "### %s\n\n", fn.Name)
			fmt.Fprintf(&buf, "```go\n%s\n```\n\n", symbol(fn))
		}
		fmt.Fprint(&buf, "\n")
	}

	return buf.String()
}

// renderSplit splits documentation by package and writes to separate files.
func renderSplit(cfg *options, defs []*model.Definition) error {
	// Create output directory if it doesn't exist
	if err := os.MkdirAll(cfg.out, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Group definitions by package
	groups := groupDefinitionsByPackage(defs)

	// Sort import paths for consistent output
	var importPaths []string
	for path := range groups {
		importPaths = append(importPaths, path)
	}
	sort.Strings(importPaths)

	// Generate documentation for each package
	var packageDocs []PackageDoc
	for _, importPath := range importPaths {
		filename := generateFilename(importPath, cfg.strip)
		content := renderMarkdownForPackage(groups[importPath])

		packageDocs = append(packageDocs, PackageDoc{
			ImportPath: importPath,
			Filename:   filename,
			Content:    content,
		})
	}

	// Write individual package documentation files
	for _, pkg := range packageDocs {
		filepath := filepath.Join(cfg.out, pkg.Filename)
		if err := os.WriteFile(filepath, []byte(pkg.Content), 0644); err != nil {
			return fmt.Errorf("failed to write %s: %w", filepath, err)
		}
	}

	// Generate README.md with table of contents
	readmePath := filepath.Join(cfg.out, "README.md")
	readmeContent := generateReadme(packageDocs)
	if err := os.WriteFile(readmePath, []byte(readmeContent), 0644); err != nil {
		return fmt.Errorf("failed to write README.md: %w", err)
	}

	return nil
}

// generateReadme creates a README.md with a table of contents.
func generateReadme(packageDocs []PackageDoc) string {
	var buf strings.Builder

	fmt.Fprint(&buf, "# API Documentation\n\n")
	fmt.Fprint(&buf, "## Table of Contents\n\n")

	for _, pkg := range packageDocs {
		// Use the import path for the section title
		fmt.Fprintf(&buf, "- [%s](./%s)\n", pkg.ImportPath, pkg.Filename)
	}

	return buf.String()
}
