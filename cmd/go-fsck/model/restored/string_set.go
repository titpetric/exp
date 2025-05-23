package model

import (
	"fmt"
	"path"
	"slices"
	"sort"
	"strings"
)

// StringSet provides a key based unique string slice.
type StringSet map[string][]string

func NewStringSet() StringSet {
	return make(StringSet)
}

func (i *StringSet) Add(key string, lits ...string) {
	data := *i
	if data == nil {
		data = make(StringSet)
	}
	if set, ok := data[key]; ok {
		for _, lit := range lits {
			if slices.Contains(set, lit) {
				return
			}
			set = append(set, lit)
		}

		data[key] = set
		return
	}
	data[key] = lits[:]
	*i = data
}

func (i StringSet) All() []string {
	result := []string{}
	for _, set := range i {
		result = append(result, set...)
	}
	return result
}

func (i StringSet) Get(key string) []string {
	val, _ := i[key]
	if val != nil {
		sort.Strings(val)
	}
	return val
}

// Map returns a map with the short package name as the key
// and the full import path as the value.
func (i StringSet) Map() (map[string]string, []error) {
	warnings := []error{}
	warningSeen := map[string]bool{}

	addWarning := func(warning error) {
		msg := warning.Error()
		if _, seen := warningSeen[msg]; !seen {
			warningSeen[msg] = true
			warnings = append(warnings, warning)
		}
	}

	cleanPackageName := func(name string) (string, bool) {
		clean := name
		clean = strings.ReplaceAll(clean, "_", "")
		return clean, name == clean
	}

	result := map[string]string{}
	imports := i.All()

	for _, imported := range imports {
		var short, long string

		// aliased package
		imported = strings.ReplaceAll(imported, "/go-", "/")
		if strings.Contains(imported, " ") {
			line := strings.Split(imported, " ")
			short, long = line[0], strings.Trim(line[1], `"`)
		} else {
			long = strings.Trim(imported, `"`)
			short = path.Base(long)
		}

		if short == "C" {
			continue
		}

		if strings.HasSuffix(short, "_test") {
			clean, ok := cleanPackageName(short[:len(short)-5])
			if !ok {
				addWarning(fmt.Errorf("Alias %s should be %s_test", short, clean))
			}
			continue
		}

		clean, ok := cleanPackageName(short)
		if !ok {
			addWarning(fmt.Errorf("Alias %s should be %s", short, clean))
			continue
		}

		val, ok := result[clean]

		if ok && val != long {
			warning := "Import conflict for %s, "
			// Sort val/long so shorter is left hand side
			if len(val) < len(long) {
				warning += val + " != " + long
			} else {
				warning += long + " != " + val
			}
			addWarning(fmt.Errorf(warning, short))
		}

		result[clean] = long
	}

	return result, warnings
}
