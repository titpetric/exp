package docs

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/titpetric/exp/cmd/go-fsck/model"
)

func TestGenerateFilename(t *testing.T) {
	tests := []struct {
		name       string
		importPath string
		strip      string
		expected   string
	}{
		{
			name:       "with strip prefix",
			importPath: "github.com/titpetric/atkins/runner/view",
			strip:      "github.com/titpetric",
			expected:   "atkins_runner_view.md",
		},
		{
			name:       "strip exact root",
			importPath: "github.com/titpetric/atkins",
			strip:      "github.com/titpetric",
			expected:   "atkins.md",
		},
		{
			name:       "without strip, exclude first element",
			importPath: "github.com/titpetric/atkins/runner",
			strip:      "",
			expected:   "titpetric_atkins_runner.md",
		},
		{
			name:       "without strip, single element",
			importPath: "mypackage",
			strip:      "",
			expected:   "mypackage.md",
		},
		{
			name:       "strip not matching",
			importPath: "github.com/other/package",
			strip:      "github.com/titpetric",
			expected:   "github.com_other_package.md",
		},
		{
			name:       "strip with trailing slash",
			importPath: "github.com/titpetric/atkins/runner",
			strip:      "github.com/titpetric/",
			expected:   "atkins_runner.md",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := generateFilename(tt.importPath, tt.strip)
			require.Equal(t, tt.expected, result)
		})
	}
}

func TestGroupDefinitionsByPackage(t *testing.T) {
	defs := []*model.Definition{
		{
			Package: model.Package{
				ImportPath: "github.com/user/pkg1",
			},
		},
		{
			Package: model.Package{
				ImportPath: "github.com/user/pkg1",
			},
		},
		{
			Package: model.Package{
				ImportPath: "github.com/user/pkg2",
			},
		},
		{
			Package: model.Package{
				ImportPath:  "github.com/user/pkg3",
				TestPackage: true,
			},
		},
	}

	groups := groupDefinitionsByPackage(defs)

	require.Len(t, groups, 2, "should have 2 groups (test package excluded)")
	require.Len(t, groups["github.com/user/pkg1"], 2, "pkg1 should have 2 definitions")
	require.Len(t, groups["github.com/user/pkg2"], 1, "pkg2 should have 1 definition")
	require.NotContains(t, groups, "github.com/user/pkg3", "test package should be excluded")
}

func TestRenderSplitBasicFunctionality(t *testing.T) {
	// Create temporary output directory
	tmpDir := t.TempDir()

	// Create minimal test definitions
	defs := []*model.Definition{
		{
			Package: model.Package{
				ImportPath: "github.com/test/pkg1",
				Path:       "./pkg1",
				Package:    "pkg1",
			},
			Doc: "Package pkg1 documentation",
		},
		{
			Package: model.Package{
				ImportPath: "github.com/test/pkg2",
				Path:       "./pkg2",
				Package:    "pkg2",
			},
			Doc: "Package pkg2 documentation",
		},
	}

	cfg := &options{
		split:  true,
		out:    tmpDir,
		strip:  "github.com/test",
		render: "markdown",
	}

	err := renderSplit(cfg, defs)
	require.NoError(t, err, "renderSplit should succeed")

	// Check that files were created
	files, err := os.ReadDir(tmpDir)
	require.NoError(t, err)

	fileNames := make(map[string]bool)
	for _, f := range files {
		fileNames[f.Name()] = true
	}

	require.True(t, fileNames["README.md"], "README.md should be created")
	require.True(t, fileNames["pkg1.md"], "pkg1.md should be created")
	require.True(t, fileNames["pkg2.md"], "pkg2.md should be created")
}

func TestRenderSplitReadmeContent(t *testing.T) {
	tmpDir := t.TempDir()

	defs := []*model.Definition{
		{
			Package: model.Package{
				ImportPath: "github.com/test/pkg1",
				Path:       "./pkg1",
				Package:    "pkg1",
			},
			Doc: "Package pkg1",
		},
		{
			Package: model.Package{
				ImportPath: "github.com/test/pkg2",
				Path:       "./pkg2",
				Package:    "pkg2",
			},
			Doc: "Package pkg2",
		},
	}

	cfg := &options{
		split:  true,
		out:    tmpDir,
		strip:  "github.com/test",
		render: "markdown",
	}

	err := renderSplit(cfg, defs)
	require.NoError(t, err)

	readmePath := filepath.Join(tmpDir, "README.md")
	content, err := os.ReadFile(readmePath)
	require.NoError(t, err)

	readmeText := string(content)
	require.Contains(t, readmeText, "# API Documentation")
	require.Contains(t, readmeText, "## Table of Contents")
	require.Contains(t, readmeText, "github.com/test/pkg1")
	require.Contains(t, readmeText, "github.com/test/pkg2")
	require.Contains(t, readmeText, "./pkg1.md")
	require.Contains(t, readmeText, "./pkg2.md")
}

func TestRenderSplitCreatesMissingDir(t *testing.T) {
	tmpDir := t.TempDir()
	outDir := filepath.Join(tmpDir, "nested", "output")

	defs := []*model.Definition{
		{
			Package: model.Package{
				ImportPath: "github.com/test/pkg1",
				Path:       "./pkg1",
				Package:    "pkg1",
			},
		},
	}

	cfg := &options{
		split:  true,
		out:    outDir,
		strip:  "github.com/test",
		render: "markdown",
	}

	err := renderSplit(cfg, defs)
	require.NoError(t, err)

	// Verify directory was created
	info, err := os.Stat(outDir)
	require.NoError(t, err)
	require.True(t, info.IsDir())
}

func TestRenderSplitWithComplexImportPath(t *testing.T) {
	tmpDir := t.TempDir()

	defs := []*model.Definition{
		{
			Package: model.Package{
				ImportPath: "github.com/titpetric/atkins/runner/view",
				Path:       "./runner/view",
				Package:    "view",
			},
		},
	}

	cfg := &options{
		split:  true,
		out:    tmpDir,
		strip:  "github.com/titpetric",
		render: "markdown",
	}

	err := renderSplit(cfg, defs)
	require.NoError(t, err)

	// Verify the correct filename was created
	expectedFile := filepath.Join(tmpDir, "atkins_runner_view.md")
	_, err = os.Stat(expectedFile)
	require.NoError(t, err, "atkins_runner_view.md should exist")
}

func TestRenderSplitExcludesTestPackages(t *testing.T) {
	tmpDir := t.TempDir()

	defs := []*model.Definition{
		{
			Package: model.Package{
				ImportPath: "github.com/test/pkg1",
				Path:       "./pkg1",
				Package:    "pkg1",
			},
		},
		{
			Package: model.Package{
				ImportPath:  "github.com/test/pkg1_test",
				Path:        "./pkg1",
				Package:     "pkg1_test",
				TestPackage: true,
			},
		},
	}

	cfg := &options{
		split:  true,
		out:    tmpDir,
		strip:  "github.com/test",
		render: "markdown",
	}

	err := renderSplit(cfg, defs)
	require.NoError(t, err)

	files, err := os.ReadDir(tmpDir)
	require.NoError(t, err)

	fileNames := make(map[string]bool)
	for _, f := range files {
		fileNames[f.Name()] = true
	}

	require.True(t, fileNames["pkg1.md"], "pkg1.md should exist")
	require.False(t, fileNames["pkg1_test.md"], "test package should not create file")
}

func TestGenerateReadmeLinks(t *testing.T) {
	packageDocs := []PackageDoc{
		{
			ImportPath: "github.com/test/pkg1",
			Filename:   "pkg1.md",
		},
		{
			ImportPath: "github.com/test/pkg2",
			Filename:   "pkg2.md",
		},
	}

	content := generateReadme(packageDocs)

	require.Contains(t, content, "# API Documentation")
	require.Contains(t, content, "## Table of Contents")
	require.Contains(t, content, "[github.com/test/pkg1](./pkg1.md)")
	require.Contains(t, content, "[github.com/test/pkg2](./pkg2.md)")
}

func TestRenderMarkdownForPackageBasic(t *testing.T) {
	defs := []*model.Definition{
		{
			Package: model.Package{
				ImportPath: "github.com/test/pkg1",
				Path:       "./pkg1",
				Package:    "pkg1",
			},
			Doc: "This is a test package",
		},
	}

	content := renderMarkdownForPackage(defs)

	require.Contains(t, content, "# Package ./pkg1")
	require.Contains(t, content, "import")
	require.Contains(t, content, "github.com/test/pkg1")
	require.Contains(t, content, "This is a test package")
}

func TestRenderMarkdownForPackageWithTypes(t *testing.T) {
	defs := []*model.Definition{
		{
			Package: model.Package{
				ImportPath: "github.com/test/pkg1",
				Path:       "./pkg1",
				Package:    "pkg1",
			},
			Types: []*model.Declaration{
				{
					Name:   "MyType",
					Source: "type MyType struct { Field string }",
				},
			},
		},
	}

	content := renderMarkdownForPackage(defs)

	require.Contains(t, content, "## Types")
	require.Contains(t, content, "MyType")
	require.Contains(t, content, "type MyType struct { Field string }")
}

func TestRenderMarkdownForPackageWithFunctions(t *testing.T) {
	defs := []*model.Definition{
		{
			Package: model.Package{
				ImportPath: "github.com/test/pkg1",
				Path:       "./pkg1",
				Package:    "pkg1",
			},
			Funcs: []*model.Declaration{
				{
					Name:      "DoSomething",
					Signature: "DoSomething(x int) string",
					Doc:       "DoSomething does something important",
					Source:    "func DoSomething(x int) string { return \"\" }",
				},
			},
		},
	}

	content := renderMarkdownForPackage(defs)

	require.Contains(t, content, "## Function symbols")
	require.Contains(t, content, "DoSomething")
	require.Contains(t, content, "DoSomething does something important")
	require.Contains(t, content, "func DoSomething(x int) string")
}

func TestRenderMarkdownForPackageEmpty(t *testing.T) {
	defs := []*model.Definition{}
	content := renderMarkdownForPackage(defs)
	require.Equal(t, "", content)
}

func TestRenderSplitMultiplePackagesOrdering(t *testing.T) {
	tmpDir := t.TempDir()

	defs := []*model.Definition{
		{
			Package: model.Package{
				ImportPath: "github.com/test/zeta",
				Path:       "./zeta",
				Package:    "zeta",
			},
		},
		{
			Package: model.Package{
				ImportPath: "github.com/test/alpha",
				Path:       "./alpha",
				Package:    "alpha",
			},
		},
		{
			Package: model.Package{
				ImportPath: "github.com/test/beta",
				Path:       "./beta",
				Package:    "beta",
			},
		},
	}

	cfg := &options{
		split:  true,
		out:    tmpDir,
		strip:  "github.com/test",
		render: "markdown",
	}

	err := renderSplit(cfg, defs)
	require.NoError(t, err)

	readmePath := filepath.Join(tmpDir, "README.md")
	content, err := os.ReadFile(readmePath)
	require.NoError(t, err)

	readmeText := string(content)
	alphaIdx := strings.Index(readmeText, "github.com/test/alpha")
	betaIdx := strings.Index(readmeText, "github.com/test/beta")
	zetaIdx := strings.Index(readmeText, "github.com/test/zeta")

	require.True(t, alphaIdx < betaIdx, "alpha should come before beta")
	require.True(t, betaIdx < zetaIdx, "beta should come before zeta")
}

func TestRenderSplitFilePermissions(t *testing.T) {
	tmpDir := t.TempDir()

	defs := []*model.Definition{
		{
			Package: model.Package{
				ImportPath: "github.com/test/pkg1",
				Path:       "./pkg1",
				Package:    "pkg1",
			},
		},
	}

	cfg := &options{
		split:  true,
		out:    tmpDir,
		strip:  "github.com/test",
		render: "markdown",
	}

	err := renderSplit(cfg, defs)
	require.NoError(t, err)

	pkgPath := filepath.Join(tmpDir, "pkg1.md")
	info, err := os.Stat(pkgPath)
	require.NoError(t, err)

	// Verify file is readable
	require.True(t, info.Mode().IsRegular())
	require.True(t, (info.Mode()&0444) != 0, "file should be readable")

	readmePath := filepath.Join(tmpDir, "README.md")
	info, err = os.Stat(readmePath)
	require.NoError(t, err)
	require.True(t, (info.Mode()&0444) != 0, "README should be readable")
}

func TestRenderSplitSpecialCharactersInPath(t *testing.T) {
	tests := []struct {
		name       string
		importPath string
		strip      string
		expected   string
	}{
		{
			name:       "long nested path",
			importPath: "github.com/titpetric/atkins/runner/view/dialog",
			strip:      "github.com/titpetric",
			expected:   "atkins_runner_view_dialog.md",
		},
		{
			name:       "double slash handling",
			importPath: "github.com/test//pkg",
			strip:      "github.com/test",
			expected:   "_pkg.md",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := generateFilename(tt.importPath, tt.strip)
			require.Equal(t, tt.expected, result)
		})
	}
}

func TestIntegrationEndToEnd(t *testing.T) {
	tmpDir := t.TempDir()

	defs := []*model.Definition{
		{
			Package: model.Package{
				ImportPath: "github.com/test/pkg1",
				Path:       "./pkg1",
				Package:    "pkg1",
			},
			Doc: "Package 1",
		},
		{
			Package: model.Package{
				ImportPath: "github.com/test/pkg2/sub",
				Path:       "./pkg2/sub",
				Package:    "sub",
			},
			Doc: "Package 2",
		},
	}

	cfg := &options{
		split: true,
		out:   tmpDir,
		strip: "github.com/test",
	}

	err := renderSplit(cfg, defs)
	require.NoError(t, err)

	// Verify files exist
	_, err = os.Stat(filepath.Join(tmpDir, "README.md"))
	require.NoError(t, err)

	_, err = os.Stat(filepath.Join(tmpDir, "pkg1.md"))
	require.NoError(t, err)

	_, err = os.Stat(filepath.Join(tmpDir, "pkg2_sub.md"))
	require.NoError(t, err)

	// Verify README content
	readmeContent, _ := os.ReadFile(filepath.Join(tmpDir, "README.md"))
	require.Contains(t, string(readmeContent), "github.com/test/pkg1")
	require.Contains(t, string(readmeContent), "github.com/test/pkg2/sub")
}
