package main

import (
	"archive/zip"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path"
	"strings"

	"github.com/olekukonko/tablewriter"
	"golang.org/x/mod/modfile"
	"golang.org/x/mod/module"
	"golang.org/x/mod/semver"
)

func main() {
	if err := start(); err != nil {
		log.Fatal(err)
	}
}

type Dependency struct {
	Name      string
	Version   string
	Latest    string
	Upgrade   bool
	Warnings  string
	FileCount int
	TotalSize int64
}

func (d *Dependency) StringSlice() []string {
	var (
		version = d.Version
		name    = d.Name
	)

	// nicer word wrap for github markdown
	version = strings.ReplaceAll(version, "-", " ")
	version = strings.ReplaceAll(version, "+", " +")

	// strip github.com for less data
	name = strings.ReplaceAll(name, "github.com/", "")

	return toStringSlice(name, version, d.Latest, formatSize(d.TotalSize), fmt.Sprintf("%d", d.FileCount), d.Warnings)
}

func loadGoMod(gomodPath string) ([]*Dependency, error) {
	content, err := os.ReadFile(gomodPath)
	if err != nil {
		return nil, err
	}
	return load(content)
}

func loadFromProxy(modulePath string) ([]*Dependency, error) {
	parts := strings.SplitN(modulePath, "@", 2)
	name := parts[0]
	version := "latest"
	if len(parts) == 2 {
		version = parts[1]
	}

	if version == "latest" {
		resolved, err := resolveLatest(name)
		if err != nil {
			return nil, fmt.Errorf("resolving latest version for %s: %w", name, err)
		}
		version = resolved
	}

	escapedName, err := module.EscapePath(name)
	if err != nil {
		return nil, fmt.Errorf("escaping module path %s: %w", name, err)
	}

	modURL := "https://proxy.golang.org/" + escapedName + "/@v/" + version + ".mod"
	res, err := http.Get(modURL)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	content, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("fetching go.mod for %s@%s: %s", name, version, strings.TrimSpace(string(content)))
	}

	return load(content)
}

func resolveLatest(name string) (string, error) {
	escapedName, err := module.EscapePath(name)
	if err != nil {
		return "", fmt.Errorf("escaping module path %s: %w", name, err)
	}
	res, err := http.Get("https://proxy.golang.org/" + escapedName + "/@latest")
	if err != nil {
		return "", err
	}
	defer res.Body.Close()

	var info struct {
		Version string `json:"Version"`
	}
	if err := json.NewDecoder(res.Body).Decode(&info); err != nil {
		return "", err
	}
	return info.Version, nil
}

func load(content []byte) ([]*Dependency, error) {
	var result []*Dependency

	f, err := modfile.ParseLax("go.mod", content, nil)
	if err != nil {
		return nil, err
	}

	// Exclude org-wide version checks
	//
	// The go module path gets stripped of the repository, so we can
	// avoid hitting proxy
	pkg := f.Module.Mod.String()
	pkg = path.Dir(pkg)

	for _, r := range f.Require {
		if r.Indirect {
			continue
		}

		fileCount, totalSize, _ := getModuleSize(r.Mod.Path, r.Mod.Version)

		dep := &Dependency{
			Name:      r.Mod.Path,
			Version:   r.Mod.Version,
			Upgrade:   false,
			Latest:    "Skipped",
			FileCount: fileCount,
			TotalSize: totalSize,
		}

		if !strings.HasPrefix(dep.Name, pkg) {
			latest, err := getLatestVersion(dep.Name)
			if err != nil {
				dep.Latest = err.Error()
			} else {
				dep.Latest = latest
			}
		}

		dep.Warnings = lintImport(dep.Name, dep.Version, dep.Latest)

		switch {
		case dep.Latest == dep.Version:
			dep.Latest = "✓ Up to date"
		case strings.HasPrefix(dep.Latest, "bad request:"):
			dep.Latest = "✖ No info"
		default:
			if dep.Warnings == "" && semver.IsValid(dep.Version) && semver.IsValid(dep.Latest) {
				if semver.Compare(dep.Version, dep.Latest) < 0 {
					dep.Upgrade = true
				}
			}
		}

		result = append(result, dep)
	}

	return result, nil
}

func isSkipped(conf *options, name string) bool {
	for _, skipped := range conf.skip {
		if strings.HasPrefix(name, skipped) {
			return true
		}
	}
	return false
}

func start() error {
	conf := NewOptions()

	var deps []*Dependency
	var err error

	args := conf.args
	if len(args) > 0 {
		deps, err = loadFromProxy(args[0])
	} else {
		deps, err = loadGoMod(conf.goModPath)
	}
	if err != nil {
		return err
	}

	// Apply skip/upgrade filters for all output modes.
	var filtered []*Dependency
	for _, dep := range deps {
		if isSkipped(conf, dep.Name) {
			dep.Warnings = "Held back from upgrade"
		}
		if conf.forUpgrade && !dep.Upgrade {
			continue
		}
		filtered = append(filtered, dep)
	}

	switch {
	case conf.json:
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(filtered)
	case conf.suggest:
		for _, dep := range filtered {
			if dep.Upgrade {
				if isSkipped(conf, dep.Name) {
					log.Println(dep.Name, "held back from upgrade")
					continue
				}
				fmt.Printf("go get %s@%s\t\t# upgrade from %s\n", dep.Name, dep.Latest, dep.Version)
			}
		}
	default:
		output := &strings.Builder{}

		w := tablewriter.NewWriter(output)
		w.SetHeader([]string{"import", "version", "latest", "size", "files", "warnings"})
		w.SetAutoWrapText(false)
		w.SetAutoFormatHeaders(true)
		w.SetTablePadding(" ")
		w.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
		w.SetAlignment(tablewriter.ALIGN_LEFT)
		w.SetRowSeparator("")
		w.SetHeaderLine(true)
		w.SetCenterSeparator("|")
		w.SetBorders(tablewriter.Border{Left: true, Top: false, Right: true, Bottom: false})

		for _, dep := range filtered {
			w.Append(dep.StringSlice())
		}

		w.Render()

		tableString := output.String()
		for strings.Contains(tableString, "||") {
			tableString = strings.Replace(tableString, "||", "|:---|", 1)
		}

		fmt.Print(tableString)
	}

	return nil
}

func getLatestVersion(name string) (string, error) {
	var result string

	escapedName, err := module.EscapePath(name)
	if err != nil {
		return result, fmt.Errorf("escaping module path %s: %w", name, err)
	}
	res, err := http.Get("https://proxy.golang.org/" + escapedName + "/@v/list")
	if err != nil {
		return result, err
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return result, err
	}

	parts := strings.Split(strings.TrimSpace(string(body)), "\n")
	cleanParts := []string{}
	for _, part := range parts {
		// Skip `-rc`, `-dev` and similar suffixes
		if strings.Contains(part, "-") {
			continue
		}
		cleanParts = append(cleanParts, part)
	}

	semver.Sort(cleanParts)

	if len(cleanParts) > 0 {
		result = cleanParts[len(cleanParts)-1]
	}
	if len(result) == 0 {
		err = errors.New("No versions available")
	}

	return result, err
}

func toStringSlice(vars ...string) []string {
	return vars
}

type modDownload struct {
	Zip string `json:"Zip"`
}

func getModuleSize(name, version string) (int, int64, error) {
	cmd := exec.Command("go", "mod", "download", "-json", name+"@"+version)
	out, err := cmd.Output()
	if err != nil {
		return 0, 0, err
	}

	var dl modDownload
	if err := json.Unmarshal(out, &dl); err != nil {
		return 0, 0, err
	}

	reader, err := zip.OpenReader(dl.Zip)
	if err != nil {
		return 0, 0, err
	}
	defer reader.Close()

	var fileCount int
	var totalSize int64
	for _, f := range reader.File {
		if f.FileInfo().IsDir() {
			continue
		}
		fileCount++
		totalSize += int64(f.UncompressedSize64)
	}

	return fileCount, totalSize, nil
}

func formatSize(size int64) string {
	switch {
	case size >= 1024*1024:
		return fmt.Sprintf("%.1f MB", float64(size)/(1024*1024))
	case size >= 1024:
		return fmt.Sprintf("%.1f KB", float64(size)/1024)
	default:
		return fmt.Sprintf("%d B", size)
	}
}

func lintImport(name, version, latest string) string {
	if strings.Contains(name, "gopkg.in/") {
		return "Deprecated import (gopkg.in)"
	}

	if strings.HasPrefix(version, "v0.0.0-") {
		return "Dependency without go.mod"
	}

	if strings.HasPrefix(latest, "bad request:") {
		return "Bad request, possibly renamed"
	}

	if latest == "Skipped" {
		return ""
	}

	versionTrimmed := strings.Split(version, "-")[0]
	if semver.Compare(versionTrimmed, latest) > 0 {
		return "Version ahead of latest release"
	}

	return ""
}
