package test

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/titpetric/exp/cmd/go-fsck/internal"
	"github.com/titpetric/exp/cmd/go-fsck/model"
)

// compileMultiModuleTests handles go test -c across multiple modules.
// It compiles test binaries for all packages, handling name conflicts
// by prefixing with the import path.
func compileMultiModuleTests(args []string) error {
	// Extract output directory from -o flag
	outputDir := "bin"
	for i, arg := range args {
		if arg == "-o" && i+1 < len(args) {
			outputDir = args[i+1]
			break
		}
	}

	// Create output directory
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// List all modules in the workspace
	modules, err := internal.ListModules(".", "./...")
	if err != nil {
		return fmt.Errorf("failed to list modules: %w", err)
	}

	if len(modules) == 0 {
		return fmt.Errorf("no go.mod files found")
	}

	// For each module, compile tests
	for _, mod := range modules {
		if !mod.Valid {
			return fmt.Errorf("invalid module %s: %w", mod.ImportPath, mod.Error)
		}

		if err := compileModuleTests(args, outputDir, mod); err != nil {
			return err
		}
	}

	fmt.Printf("Successfully compiled test binaries to %s/\n", outputDir)
	return nil
}

// compileModuleTests compiles tests for a single module.
func compileModuleTests(args []string, outputDir string, mod internal.Module) error {
	// List all packages first to determine which ones to try compiling
	// We use the internal ListPackages which loads with Tests: true
	allPackages, err := internal.ListPackages(mod.Dir, "./...")
	if err != nil {
		return fmt.Errorf("failed to list packages in module %s: %w", mod.ImportPath, err)
	}

	// Try to compile each non-synthetic package
	// When packages.Load is used with Tests: true, it returns synthetic test packages too
	// We compile all non-synthetic packages and skip those without tests
	packagesToTry := make(map[string]*model.Package)
	for _, pkg := range allPackages {
		// Skip synthetic test packages (.test and _test suffixes)
		if strings.HasSuffix(pkg.ImportPath, ".test") || strings.HasSuffix(pkg.ImportPath, "_test") {
			continue
		}
		// Use all non-synthetic packages as candidates for compilation
		if _, exists := packagesToTry[pkg.ImportPath]; !exists {
			packagesToTry[pkg.ImportPath] = pkg
		}
	}

	if len(packagesToTry) == 0 {
		fmt.Printf("No packages found in module %s\n", mod.ImportPath)
		return nil
	}

	// Try to compile each package, gracefully skip those without test files
	compiledCount := 0
	for _, pkg := range packagesToTry {
		if err := compilePackageTest(args, outputDir, mod.Dir, pkg); err != nil {
			// Skip packages that don't have test files
			if strings.Contains(err.Error(), "no non-test") || strings.Contains(err.Error(), "no Go files") {
				continue
			}
			return err
		}
		compiledCount++
	}

	if compiledCount == 0 {
		fmt.Printf("No test packages found in module %s\n", mod.ImportPath)
	}

	return nil
}

// compilePackageTest compiles a single package's tests into a uniquely named binary.
func compilePackageTest(args []string, outputDir, moduleDir string, pkg *model.Package) error {
	// Create a deterministic output filename to avoid collisions
	// Use the full package path with slashes replaced by underscores
	outName := fmt.Sprintf("%s.test", strings.ReplaceAll(pkg.ImportPath, "/", "_"))
	outPath := filepath.Join(outputDir, outName)

	// Build the go test -c command for this package only
	cmdArgs := []string{"test", "-c"}

	// Add all pass-through args except package patterns and -o output
	skipNext := false
	for _, arg := range args {
		if skipNext {
			skipNext = false
			continue
		}
		if arg == "-o" {
			skipNext = true
			continue
		}
		if !isPackagePattern(arg) {
			cmdArgs = append(cmdArgs, arg)
		}
	}

	// Add output flag - for single package compilation, use full file path
	cmdArgs = append(cmdArgs, "-o", outPath)

	// Add the package to test
	cmdArgs = append(cmdArgs, pkg.ImportPath)

	// Run go test from the module directory
	cmd := exec.Command("go", cmdArgs...)
	cmd.Dir = moduleDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to compile tests for %s: %w", pkg.ImportPath, err)
	}

	fmt.Printf("Compiled: %s\n", outPath)
	return nil
}

// isPackagePattern checks if an argument is a package pattern rather than a flag value.
func isPackagePattern(arg string) bool {
	// Common patterns that indicate package specifications
	if arg == "./..." || arg == "..." || strings.HasPrefix(arg, "./") ||
		strings.HasPrefix(arg, "../") || strings.Contains(arg, "/...") {
		return true
	}
	// Single dots that are paths
	if arg == "." {
		return true
	}
	return false
}
