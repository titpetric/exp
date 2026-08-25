package diff

import (
	"encoding/json"
	"fmt"
	"sort"

	"github.com/titpetric/exp/cmd/go-fsck/model"
	"github.com/titpetric/exp/cmd/go-fsck/model/loader"
)

// Symbol is one exported declaration of a package.
type Symbol struct {
	// Key identifies the declaration across revisions, as the import path
	// followed by the receiver type and name.
	Key string `json:"key"`

	// Package is the import path holding the declaration.
	Package string `json:"package"`

	// Name is the declared name, qualified with its receiver type when it
	// has one, as in "Manager.Start".
	Name string `json:"name"`

	// Kind is type, const, var or func.
	Kind string `json:"kind"`

	// Signature is a func as it is declared, and is empty for every other
	// kind, which is named by Kind and Name alone.
	Signature string `json:"signature,omitempty"`

	// Definition is a type as it is declared, without its doc comment. It is
	// empty for every other kind, and for a type whenever the model was
	// extracted without sources.
	Definition string `json:"definition,omitempty"`
}

// String renders the declaration the way it reads in source.
func (s Symbol) String() string {
	if s.Signature != "" {
		return s.Signature
	}
	return s.Kind + " " + s.Name
}

// Change is an exported symbol whose signature moved between two revisions.
type Change struct {
	Key     string `json:"key"`
	Package string `json:"package"`
	Name    string `json:"name"`

	// Old and New are the signature before and after, with parameter names
	// removed, which is what they were compared on.
	Old string `json:"old"`
	New string `json:"new"`
}

// Result is the exported API difference between two revisions of a module.
type Result struct {
	// Removed holds the symbols the older revision had and the newer one
	// does not.
	Removed []Symbol `json:"removed"`

	// Added holds the symbols the newer revision gained.
	Added []Symbol `json:"added"`

	// Changed holds the symbols both revisions have under a signature that
	// moved.
	Changed []Change `json:"changed"`

	// Breaking reports whether the difference takes API away, which is a
	// removed symbol or a changed signature. Added symbols are not breaking.
	Breaking bool `json:"breaking"`
}

// Compare returns the exported API difference between two sets of definitions.
// Symbols are keyed by import path, receiver type and name, so a declaration
// moving to another file, or its group gaining a sibling, is not a difference.
func Compare(oldDefs, newDefs []*model.Definition, includeInternal bool) Result {
	var (
		old = symbols(oldDefs, includeInternal)
		cur = symbols(newDefs, includeInternal)
	)

	result := Result{
		Removed: []Symbol{},
		Added:   []Symbol{},
		Changed: []Change{},
	}

	for key, was := range old {
		is, ok := cur[key]
		switch {
		case !ok:
			result.Removed = append(result.Removed, was.symbol)
		case is.normalized != was.normalized:
			result.Changed = append(result.Changed, Change{
				Key:     key,
				Package: is.symbol.Package,
				Name:    is.symbol.Name,
				Old:     was.normalized,
				New:     is.normalized,
			})
		}
	}
	for key, is := range cur {
		if _, ok := old[key]; !ok {
			result.Added = append(result.Added, is.symbol)
		}
	}

	// The maps are walked in whatever order the runtime hands out, and the
	// declaration order of the input is not dependable either, so the lists
	// are sorted to make the output of two identical runs identical.
	sortSymbols(result.Removed)
	sortSymbols(result.Added)
	sort.Slice(result.Changed, func(i, j int) bool { return result.Changed[i].Key < result.Changed[j].Key })

	result.Breaking = len(result.Removed) > 0 || len(result.Changed) > 0
	return result
}

func sortSymbols(symbols []Symbol) {
	sort.Slice(symbols, func(i, j int) bool { return symbols[i].Key < symbols[j].Key })
}

func diff(cfg *options) error {
	oldDefs, err := loader.ReadFile(cfg.oldFile)
	if err != nil {
		return fmt.Errorf("read %s: %w", cfg.oldFile, err)
	}
	newDefs, err := loader.ReadFile(cfg.newFile)
	if err != nil {
		return fmt.Errorf("read %s: %w", cfg.newFile, err)
	}

	result := Compare(oldDefs, newDefs, cfg.includeInternal)

	if cfg.json {
		b, err := json.Marshal(result)
		if err != nil {
			return err
		}
		fmt.Println(string(b))
		return nil
	}

	for _, symbol := range result.Removed {
		fmt.Printf("- %s\n", symbol.Key)
		if cfg.verbose {
			fmt.Printf("    %s\n", symbol)
		}
	}
	for _, change := range result.Changed {
		fmt.Printf("~ %s\n", change.Key)
		if cfg.verbose {
			fmt.Printf("    %s\n    %s\n", change.Old, change.New)
		}
	}
	for _, symbol := range result.Added {
		fmt.Printf("+ %s\n", symbol.Key)
		if cfg.verbose {
			fmt.Printf("    %s\n", symbol)
		}
	}

	summary := fmt.Sprintf("%d removed, %d changed, %d added", len(result.Removed), len(result.Changed), len(result.Added))
	if result.Breaking {
		summary += ", breaking"
	}
	fmt.Println(summary)
	return nil
}
