package internal

import (
	"fmt"
	"log"
	"os"
	"strings"

	"golang.org/x/tools/go/packages"

	"github.com/titpetric/exp/cmd/go-fsck/model"
)

// ListPackages returns a slice of local packages in the specified root directory.
// The second argument is either `.` or `./...` (recursive).
func ListPackages(rootPath string, pattern string) ([]*model.Package, error) {
	if err := os.Chdir(rootPath); err != nil {
		return nil, err
	}

	cwd, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	packages, err := listPackages(pattern)
	if err != nil || len(packages) == 0 {
		return nil, err
	}

	result := cleanPackages(packages, cwd)

	return result, nil
}

func cleanPackages(pkgs []*packages.Package, workDir string) []*model.Package {
	results := make([]*model.Package, 0, len(pkgs))

	isDebug := false
	if isDebug {
		// This area is sensitive. ID combines test context and package names.
		for _, pkg := range pkgs {
			fmt.Printf("- %s [%s, %q]\n", pkg.Name, pkg.Dir, pkg.ID)
		}
	}

	for _, pkg := range pkgs {
		if isDebug {
			log.Printf("> %q, %q %q %q", pkg.ID, pkg.Name, pkg.PkgPath, pkg.ForTest)
			for _, f := range pkg.GoFiles {
				log.Println("-", f)
			}
		}

		isTestScope := false
		for _, f := range pkg.GoFiles {
			if strings.HasSuffix(f, "_test.go") {
				isTestScope = true
			}
		}

		result := &model.Package{
			ID:          pkg.ID,
			Package:     pkg.Name, //filepath.Base(pkg.PkgPath),
			ImportPath:  pkg.PkgPath,
			Path:        "." + strings.TrimPrefix(pkg.Dir, workDir),
			TestPackage: isTestScope,
			Pkg:         pkg,
		}

		results = append(results, result)
	}

	if isDebug {
		fmt.Println("Done with", len(results))
	}

	return results
}

func listPackages(pattern string) ([]*packages.Package, error) {
	cfg := &packages.Config{
		Mode:  packages.NeedName | packages.NeedFiles | packages.NeedCompiledGoFiles | packages.NeedModule | packages.LoadAllSyntax,
		Tests: true,
	}

	return packages.Load(cfg, pattern)
}
