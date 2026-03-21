package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
)

func findGoWork() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for {
		p := filepath.Join(dir, "go.work")
		if _, err := os.Stat(p); err == nil {
			return p, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return "", os.ErrNotExist
}

func main() {
	update := flag.Bool("u", false, "update workspace dependencies to latest tags")
	flag.Parse()

	goWorkPath, err := findGoWork()
	if err != nil {
		log.Fatalf("go.work not found in current or parent directories")
	}
	if err := os.Chdir(filepath.Dir(goWorkPath)); err != nil {
		log.Fatalf("failed to chdir to %s: %v", filepath.Dir(goWorkPath), err)
	}

	modDirs, err := parseGoWork("go.work")
	if err != nil {
		log.Fatalf("failed to parse go.work: %v", err)
	}

	// Map: module path -> dir
	modPaths := make(map[string]string)
	for _, dir := range modDirs {
		modPath, err := readModulePath(dir)
		if err != nil {
			log.Fatalf("failed to read module in %s: %v", dir, err)
		}
		modPaths[modPath] = dir
	}

	// Build dependency map (uses) and version map
	uses := make(map[string][]string)
	versionRefs := make(map[string]map[string]string)
	for modPath, dir := range modPaths {
		reqs, err := readRequiresVersioned(dir)
		if err != nil {
			log.Fatalf("failed to read requires for %s: %v", modPath, err)
		}
		for _, r := range reqs {
			if _, ok := modPaths[r.path]; ok {
				uses[modPath] = append(uses[modPath], r.path)
				if versionRefs[modPath] == nil {
					versionRefs[modPath] = make(map[string]string)
				}
				versionRefs[modPath][r.path] = r.version
			}
		}
	}

	// Build reverse map (used_by)
	usedBy := make(map[string][]string)
	for mod, deps := range uses {
		for _, dep := range deps {
			usedBy[dep] = append(usedBy[dep], mod)
		}
	}

	// Get latest git tag for each module
	latestTags := make(map[string]string)
	for modPath, dir := range modPaths {
		tag := latestGitTag(dir)
		if tag != "" {
			latestTags[modPath] = tag
		}
	}

	// Get git status for each module
	gitStatuses := make(map[string]*gitStatus)
	for modPath, dir := range modPaths {
		if st := getGitStatus(dir); st != nil {
			gitStatuses[modPath] = st
		}
	}

	// Build sorted output: order by count(used_by) desc, count(uses) asc, name asc
	var sortedMods []string
	for mod := range modPaths {
		sortedMods = append(sortedMods, mod)
	}
	sort.Slice(sortedMods, func(i, j int) bool {
		ubi, ubj := len(usedBy[sortedMods[i]]), len(usedBy[sortedMods[j]])
		if ubi != ubj {
			return ubi > ubj
		}
		ui, uj := len(uses[sortedMods[i]]), len(uses[sortedMods[j]])
		if ui != uj {
			return ui < uj
		}
		return sortedMods[i] < sortedMods[j]
	})

	// Build module info list
	var modules []moduleInfo
	for _, mod := range sortedMods {
		info := moduleInfo{Name: mod}
		if tag, ok := latestTags[mod]; ok {
			info.Latest = tag
			info.Ahead = commitsSinceTag(modPaths[mod], tag)
		}
		if st, ok := gitStatuses[mod]; ok {
			info.Git = formatGitSummary(st)
		} else if info.Ahead > 0 {
			info.GitMsgs = commitMessagesSinceTag(modPaths[mod], latestTags[mod])
		}
		if deps, ok := uses[mod]; ok {
			sort.Strings(deps)
			info.Uses = deps
		}
		if revs, ok := usedBy[mod]; ok {
			sort.Strings(revs)
			info.UsedBy = revs
		}
		modules = append(modules, info)
	}

	if *update {
		updateDeps(modPaths, versionRefs, latestTags)
		return
	}

	renderTables(modules, versionRefs, latestTags, gitStatuses)
}

func updateDeps(modPaths map[string]string, versionRefs map[string]map[string]string, latestTags map[string]string) {
	for modPath, refs := range versionRefs {
		dir := modPaths[modPath]
		modShort := filepath.Base(modPath)
		updated := false

		for dep, ver := range refs {
			latest := latestTags[dep]
			if latest == "" || ver == latest {
				continue
			}
			depShort := filepath.Base(dep)
			fmt.Printf("Updated %s %s@%s to %s@%s\n", modShort, depShort, ver, depShort, latest)

			cmd := exec.Command("go", "get", dep+"@"+latest)
			cmd.Dir = dir
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			if err := cmd.Run(); err != nil {
				log.Printf("  go get failed for %s in %s: %v", dep, modPath, err)
			}
			updated = true
		}

		if updated {
			cmd := exec.Command("go", "get", "-u", "./...")
			cmd.Dir = dir
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			if err := cmd.Run(); err != nil {
				log.Printf("  go get -u failed in %s: %v", modPath, err)
			}

			cmd = exec.Command("go", "mod", "tidy")
			cmd.Dir = dir
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			if err := cmd.Run(); err != nil {
				log.Printf("  go mod tidy failed in %s: %v", modPath, err)
			}
		}
	}
}
