package list

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"golang.org/x/exp/slices"

	"github.com/titpetric/exp/cmd/schema-gen/model"
	. "github.com/titpetric/exp/cmd/schema-gen/model"
)

func listStructures(cfg *options) error {
	matches, err := filepath.Glob(cfg.inputFile)
	if err != nil {
		return err
	}

	sort.Strings(matches)

	all := []*TypeDeclaration{}

	find := func(ts []*TypeDeclaration, p string) *TypeDeclaration {
		for _, t := range ts {
			if t.Path == p {
				return t
			}
		}
		return nil
	}

	for _, filename := range matches {
		pkgInfos, err := model.Load(filename)
		if err != nil {
			return fmt.Errorf("Error loading package info for %s: %w", filename, err)
		}

		for _, pkgInfo := range pkgInfos {
			pkg := listSymbols(cfg, pkgInfo)
			for _, sym := range pkg {
				got := find(all, sym.Path)
				if got != nil {
					got.Doc = sym.Doc
					got.Type = sym.Type
					got.AddedFiles = append(got.AddedFiles, filename)
					continue
				}
				sym.AddedFiles = []string{filename}
				all = append(all, sym)
			}
		}

		for _, sym := range all {
			if len(sym.AddedFiles) > 0 {
				isRemoved := !slices.Contains(sym.AddedFiles, filename)
				if isRemoved {
					sym.RemovedFiles = append(sym.RemovedFiles, filename)
				}
			}
		}
	}

	for _, sym := range all {
		sym.Added = sym.GetVersions(sym.AddedFiles)
		sym.Removed = sym.GetVersions(sym.RemovedFiles)

		sym.Removed = SanitizeSet(sym.Added, sym.Removed)
	}

	return printSymbols(cfg, all)
}

type TypeDeclaration struct {
	AddedFiles   []string `json:"-"`
	RemovedFiles []string `json:"-"`

	Added   []string `json:"added"`
	Removed []string `json:"removed,omitempty"`

	Name     string `json:"name"`
	Path     string `json:"path"`
	Type     string `json:"type"`
	Tag      string `json:"tag"`
	JSONName string `json:"json_name"`
	Doc      string `json:"doc"`
}

var (
	majorVer = regexp.MustCompile(`v([0-9]+\.0\.0)`)
	minorVer = regexp.MustCompile(`v([0-9]+\.[0-9]+)\.`)
	patchVer = regexp.MustCompile(`v([0-9]+\.[0-9]+\.[0-9]+)`)
)

// GetVersions extracts version info from Seen. It reports the first
// minor versions where a type declaration has been seen.
func (t *TypeDeclaration) GetVersions(from []string) []string {
	majors := []string{}
	minors := []string{}
	patches := []string{}

	for _, seen := range from {
		patch := patchVer.FindString(seen)
		minor := minorVer.FindString(seen)

		major := majorVer.FindString(seen)
		if major != "" {
			parts := strings.Split(major, ".")
			major = parts[0]
			majors = append(majors, major)
			minors = append(minors, minor)
			patches = append(patches, patch)
		} else {
			parts := strings.Split(patch, ".")
			major = parts[0]
		}

		// If a field exists in vX.0.0 don't list more
		// versions after that version.
		if slices.Contains(majors, major) {
			continue
		}

		if !slices.Contains(minors, minor) {
			minors = append(minors, minor)
			patches = append(patches, patch)
		}
	}

	if len(patches) > 0 && strings.HasSuffix(patches[0], ".0") {
		return []string{patches[0]}
	}

	return SanitizeList(patches)
}

func (t *TypeDeclaration) String() string {
	var result string
	if t.Tag != "" {
		result = t.Path + " " + t.Type + " " + t.Tag
	}
	result = t.Path + " " + t.Type

	if len(t.Added) > 0 {
		result += " added " + fmt.Sprint(t.Added)
		if len(t.Removed) > 0 {
			result += " removed " + fmt.Sprint(t.Removed)
		}
	}

	return result
}

// PackageFileMap key is symbol Path for the struct ordered into a file.
type PackageFileMap map[string]*TypeDeclaration

func printSymbols(cfg *options, symbols []*TypeDeclaration) error {
	if cfg.json || cfg.prettyJSON {
		var (
			out []byte
			err error
		)
		if cfg.prettyJSON {
			out, err = json.MarshalIndent(symbols, "", "  ")
		} else {
			out, err = json.Marshal(symbols)
		}
		if err != nil {
			return err
		}

		fmt.Println(string(out))
		return nil
	}

	for _, symbol := range symbols {
		fmt.Println(symbol)
	}
	return nil
}

func listSymbols(cfg *options, pkgInfo *PackageInfo) []*TypeDeclaration {
	files := []*TypeDeclaration{}

	for _, decls := range pkgInfo.Declarations {
		for _, typeDecl := range decls.Types {
			for _, field := range typeDecl.Fields {
				files = append(files, &TypeDeclaration{
					Name:     field.Name,
					Path:     field.Path,
					Type:     field.Type,
					Tag:      field.Tag,
					Doc:      field.Doc,
					JSONName: field.JSONName,
				})
			}
		}
	}

	if cfg.sorted {
		sort.Slice(files, func(i, j int) bool {
			return files[i].Path < files[j].Path
		})
	}

	return files
}
