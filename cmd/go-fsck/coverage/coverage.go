package coverage

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sort"
	"strings"
	"text/template"

	"github.com/fbiville/markdown-table-formatter/pkg/markdown"

	"github.com/titpetric/exp/cmd/go-fsck/internal"
	"github.com/titpetric/exp/cmd/go-fsck/model"
	"github.com/titpetric/exp/cmd/go-fsck/model/loader"
)

func getDefinitions(cfg *options) ([]*model.Definition, error) {
	// Read the exported go-fsck.json data.
	defs, err := loader.ReadFile(cfg.inputFile)
	if err == nil {
		return defs, nil
	}

	// list current local packages
	packages, err := internal.ListPackages(".", ".")
	if err != nil {
		return nil, err
	}

	defs = []*model.Definition{}

	for _, pkg := range packages {
		d, err := loader.Load(pkg, false, cfg.verbose)
		if err != nil {
			return nil, err
		}

		defs = append(defs, d...)
	}

	return defs, nil
}

var (
	coveragePass = fmt.Sprintf("%c", '\U00002705')
	coverageFail = fmt.Sprintf("%c", '\U0000274C')
)

type CoverageInfo struct {
	// copy of github.com/titpetric/exp/cmd/summary/coverfunc.CoverageInfo
	File      string `json:",omitempty"`
	Filename  string `json:",omitempty"`
	Package   string
	Line      int    `json:",omitempty"`
	Function  string `json:",omitempty"`
	Functions int    `json:",omitempty"`
	Coverage  float64

	Cognitive int
}

type Coverage struct {
	Files     []CoverageInfo
	Packages  []CoverageInfo
	Functions []CoverageInfo
}

func loadCoverage(name string) (*Coverage, error) {
	result := &Coverage{}

	// We can just print coverage.
	if name == "" {
		return result, nil
	}

	b, err := os.ReadFile(name)
	if err != nil {
		return nil, fmt.Errorf("Error reading %s: %w", name, err)
	}

	if err := json.Unmarshal(b, &result); err != nil {
		return nil, fmt.Errorf("Error decoding %s: %w", name, err)
	}
	return result, err
}

func coverage(cfg *options) error {
	defs, err := getDefinitions(cfg)
	if err != nil {
		return err
	}

	coverinfo, err := loadCoverage(cfg.coverageFile)
	if err != nil {
		return err
	}

	findPackage := func(defs []*model.Definition, name string) *model.Definition {
		for _, def := range defs {
			if name == def.Package.ImportPath {
				return def
			}
		}
		return nil
	}

	var total, skipped, inits int
	for _, info := range coverinfo.Functions {
		p := findPackage(defs, info.Package)
		if p == nil {
			log.Println("Warning, can't find package %s, skipping", info.Package)
			continue
		}

		// init may show up multiple times in one package
		if info.Function == "init" {
			inits++
			p.InitCount++
			continue
		}

		f := p.Funcs.Find(func(d *model.Declaration) bool {
			if d.Kind != model.FuncKind {
				return false
			}
			return d.Name == info.Function && d.File == info.File && d.Line == info.Line
		})
		if f == nil {
			skipped++
			continue
			// return fmt.Errorf("Can't find function by name: %v", info)
		}

		if f.Complexity != nil {
			f.Complexity.Coverage = info.Coverage
		} else {
			f.Complexity = &model.Complexity{
				Coverage: info.Coverage,
			}
		}
		if info.Coverage > 0 {
			total++
		}
	}

	var totalPackages int
	for _, info := range coverinfo.Packages {
		p := findPackage(defs, info.Package)
		if p == nil {
			return fmt.Errorf("Can't find package by name: %s", info.Package)
		}

		if p.Complexity != nil {
			p.Complexity.Coverage = info.Coverage
		} else {
			p.Complexity = &model.Complexity{
				Coverage: info.Coverage,
			}
		}

		p.Funcs.Walk(func(d *model.Declaration) {
			if d.Kind != model.FuncKind {
				return
			}
			p.Complexity.Cognitive += d.Complexity.Cognitive
			p.Complexity.Cyclomatic += d.Complexity.Cyclomatic
			p.Complexity.Lines += d.Complexity.Lines
		})

		if info.Coverage > 0 {
			totalPackages++
		}
	}

	if cfg.outputFile != "" {
		b, err := json.MarshalIndent(defs, "", "  ")
		if err != nil {
			return err
		}

		if err := os.WriteFile(cfg.outputFile, b, 0644); err != nil {
			return err
		}

		fmt.Printf("Wrote function coverage %d/%d (skipped %d funcs, %d init()), package coverage %d/%d to %s\n", total, len(coverinfo.Functions), skipped, inits, totalPackages, len(coverinfo.Packages), cfg.outputFile)
	} else {
		packages := func(defs []*model.Definition) []model.Package {
			var result []model.Package
			for _, def := range defs {
				result = append(result, def.Package)
			}
			sort.Slice(result, func(i, j int) bool {
				var k, v = result[i], result[j]
				if k.ImportPath != v.ImportPath {
					return strings.Compare(k.ImportPath, v.ImportPath) < 0
				}
				return false
			})
			return result
		}(defs)

		var result []CoverageInfo
		for _, def := range defs {
			fns := def.Funcs.Filter(func(d *model.Declaration) bool {
				if cfg.verbose {
					return true
				}
				return d.Complexity != nil && d.Complexity.Coverage > 0
			})
			for _, fn := range fns {
				info := CoverageInfo{
					Package:   def.Package.ImportPath,
					Function:  combined(fn.Receiver, fn.Name),
					Coverage:  fn.Complexity.Coverage,
					Cognitive: fn.Complexity.Cognitive,
				}
				result = append(result, info)
			}
		}

		sort.Slice(result, func(i, j int) bool {
			var k, v = result[i], result[j]
			if k.Package != v.Package {
				return strings.Compare(k.Package, v.Package) < 0
			}
			if k.Function != v.Function {
				return strings.Compare(k.Function, v.Function) < 0
			}
			return k.Coverage > v.Coverage
		})

		if cfg.json {
			encoder := json.NewEncoder(os.Stdout)
			encoder.SetIndent("", "  ")
			return encoder.Encode(result)
		}

		type templateData struct {
			Packages  string
			Functions string
		}
		data := &templateData{}

		{
			vars := [][]string{}
			for _, r := range result {
				coverage := r.Coverage
				cognit := r.Cognitive
				pass := cognit == 0 || coverage > 80 || (coverage > 0 && cognit <= 5)
				passText := coveragePass
				if !pass {
					passText = coverageFail
				}

				vars = append(vars, []string{passText, r.Package, r.Function, fmt.Sprintf("%.2f%%", coverage), fmt.Sprint(cognit)})
			}

			table, err := markdown.NewTableFormatterBuilder().WithPrettyPrint().Build("Status", "Package", "Function", "Coverage", "Cognitive").Format(vars)
			if err != nil {
				return err
			}
			data.Functions = strings.TrimSpace(fmt.Sprint(table))
		}

		{
			vars := [][]string{}

			for _, r := range packages {
				if r.Complexity == nil {
					r.Complexity = &model.Complexity{}
				}
				coverage := r.Complexity.Coverage
				cognit := r.Complexity.Cognitive
				lines := r.Complexity.Lines
				pass := cognit == 0 || coverage > 80 || (coverage > 0 && cognit <= 5)
				passText := coveragePass
				if !pass {
					passText = coverageFail
				}

				vars = append(vars, []string{passText, r.ImportPath, fmt.Sprintf("%.2f%%", coverage), fmt.Sprint(cognit), fmt.Sprint(lines)})
			}

			table, err := markdown.NewTableFormatterBuilder().WithPrettyPrint().Build("Status", "Package", "Coverage", "Cognitive", "Lines").Format(vars)
			if err != nil {
				return err
			}
			data.Packages = strings.TrimSpace(fmt.Sprint(table))
		}

		if cfg.template != "" {
			tmpl, err := template.ParseFiles(cfg.template)
			if err != nil {
				return err
			}

			var buf bytes.Buffer
			if err := tmpl.Execute(&buf, data); err != nil {
				return err
			}

			fmt.Println(buf.String())
		} else {
			fmt.Println(data.Functions)
		}
	}

	return nil
}

func combined(receiver, name string) string {
	if receiver != "" {
		return strings.TrimLeft(receiver, "*") + "." + name
	}
	return name
}
