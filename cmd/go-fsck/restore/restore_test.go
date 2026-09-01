package restore

import (
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/titpetric/exp/cmd/go-fsck/internal"
	"github.com/titpetric/tools/splint"
	"github.com/titpetric/tools/splint/loader"
	"github.com/titpetric/tools/splint/model"
)

// source is the package the round trip is run over: a type with a method and a
// constructor, an unexported helper, a value each way, an import one file
// reaches and the other does not, and a test.
var source = map[string]string{
	"go.mod": "module example.com/fixture\n\ngo 1.24.0\n",

	"trace.go": `// Package fixture is restored and built again.
package fixture

import (
	"fmt"
	"time"
)

// Trace is a recorded trace.
type Trace struct {
	Name  string
	Start time.Time
}

// NewTrace returns a trace named for what it records.
func NewTrace(name string) *Trace {
	return &Trace{Name: sanitise(name), Start: time.Now()}
}

// String names the trace the way a reader reads it.
func (t *Trace) String() string {
	return fmt.Sprintf("%s at %s", t.Name, t.Start)
}
`,

	"util.go": `package fixture

import "strings"

// DefaultName is what an unnamed trace is called.
const DefaultName = "anonymous"

// sanitise trims what a name is written with.
func sanitise(name string) string {
	if name = strings.TrimSpace(name); name == "" {
		return DefaultName
	}
	return name
}
`,

	"trace_test.go": `package fixture

import "testing"

func TestTrace(t *testing.T) {
	if NewTrace("  ").Name != DefaultName {
		t.Error("an unnamed trace is not named")
	}
}
`,
}

// write lays the fixture out in a directory of its own.
func write(t *testing.T, files map[string]string) string {
	t.Helper()

	dir := t.TempDir()
	for name, body := range files {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return dir
}

// TestRestoreRoundTrip covers the claim the command makes: a package extracted
// and restored is a package that builds and passes its own tests.
func TestRestoreRoundTrip(t *testing.T) {
	if _, err := exec.LookPath("go"); err != nil {
		t.Skip("the go toolchain is not on the path")
	}

	from := write(t, source)
	document := filepath.Join(t.TempDir(), "go-fsck.json")

	defs, err := internal.Definitions(splint.Options{
		SourcePath:     from,
		Pattern:        ".",
		IncludeSources: true,
		IncludeTests:   true,
	})
	if err != nil {
		t.Fatalf("reading the fixture: %v", err)
	}
	if err := loader.Save(document, &model.DocumentRoot{SchemaVersion: model.SchemaVersion, Packages: defs}); err != nil {
		t.Fatalf("writing the document: %v", err)
	}

	into := t.TempDir()
	cfg := &options{inputFile: document, outputPath: into}
	if err := restore(cfg); err != nil {
		t.Fatalf("restore() error = %v", err)
	}

	// Every symbol is in the file named for it, the unexported half is in the
	// file named for the package, the values with no type of their own are
	// together, and the test is beside what it covers.
	want := []string{"const.go", "doc.go", "fixture.go", "trace.go", "trace_test.go"}
	if got := written(t, into); !sameList(got, want) {
		t.Errorf("restore wrote %v, want %v", got, want)
	}

	// The imports of a file are the ones its declarations reach: trace.go
	// reaches time and fmt, and strings is written where sanitise is.
	body := read(t, filepath.Join(into, "trace.go"))
	for _, want := range []string{`"fmt"`, `"time"`} {
		if !strings.Contains(body, want) {
			t.Errorf("trace.go does not import %s:\n%s", want, body)
		}
	}
	if strings.Contains(body, `"strings"`) {
		t.Errorf("trace.go imports strings, which nothing in it reaches:\n%s", body)
	}

	if err := os.WriteFile(filepath.Join(into, "go.mod"), []byte(source["go.mod"]), 0o644); err != nil {
		t.Fatal(err)
	}

	out, err := run(into, "go", "test", "./...")
	if err != nil {
		t.Fatalf("the restored package does not pass its tests: %v\n%s", err, out)
	}
}

// TestRestoreWithoutSources covers the document extracted without them: there
// is nothing to write, and the command says which flag was missing.
func TestRestoreWithoutSources(t *testing.T) {
	from := write(t, source)
	document := filepath.Join(t.TempDir(), "go-fsck.json")

	defs, err := internal.Definitions(splint.Options{SourcePath: from, Pattern: "."})
	if err != nil {
		t.Fatalf("reading the fixture: %v", err)
	}
	if err := loader.Save(document, &model.DocumentRoot{SchemaVersion: model.SchemaVersion, Packages: defs}); err != nil {
		t.Fatal(err)
	}

	err = restore(&options{inputFile: document, outputPath: t.TempDir()})
	if err == nil {
		t.Fatal("restore() wrote a package with no sources in it")
	}
	if !strings.Contains(err.Error(), "--include-sources") {
		t.Errorf("error = %v, want the flag that was missing", err)
	}
}

// written are the files a restore produced, in order.
func written(t *testing.T, dir string) []string {
	t.Helper()

	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}

	var out []string
	for _, entry := range entries {
		out = append(out, entry.Name())
	}
	sort.Strings(out)
	return out
}

func read(t *testing.T, name string) string {
	t.Helper()

	body, err := os.ReadFile(name)
	if err != nil {
		t.Fatal(err)
	}
	return string(body)
}

func run(dir, name string, args ...string) (string, error) {
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	return string(out), err
}

func sameList(got, want []string) bool {
	if len(got) != len(want) {
		return false
	}
	for i := range got {
		if got[i] != want[i] {
			return false
		}
	}
	return true
}
