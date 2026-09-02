package main

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/titpetric/exp/cmd/go-ddd-stats/model"
)

// testStats is one file in each of two buckets, with a third bucket empty.
func testStats() *model.Stats {
	stats := &model.Stats{}
	stats.AppendFile(&model.File{Name: "a.go", Package: "main", Size: 500})
	stats.AppendFile(&model.File{Name: "b.go", Package: "main", Size: 1500})
	stats.AppendFile(&model.File{Name: "c.go", Package: "main", Size: 1600})
	stats.Histogram = model.Histogram(stats)
	return stats
}

// TestRenderD2 pins what the d2 binary is handed: a title naming the file
// count, one node per bucket that holds files, and the arrows between them.
func TestRenderD2(t *testing.T) {
	var out strings.Builder
	if err := renderD2(&out, testStats()); err != nil {
		t.Fatalf("renderD2: %v", err)
	}
	got := out.String()

	for _, want := range []string{
		"File size distribution (*.go, 3 files)",
		`b0: "< 1 KB"`,
		"count: 1",
		`b1: "< 2 KB"`,
		"count: 2",
		"b0 -> b1",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("diagram is missing %q:\n%s", want, got)
		}
	}
	// An empty bucket is a gap in the range, not a measurement.
	if strings.Contains(got, "< 4 KB") {
		t.Errorf("an empty bucket reached the diagram:\n%s", got)
	}
}

// TestRenderStats pins the summary: the histogram and the counts, without the
// files behind them.
func TestRenderStats(t *testing.T) {
	var out strings.Builder
	if err := renderStats(&out, testStats()); err != nil {
		t.Fatalf("renderStats: %v", err)
	}
	var report struct {
		Sizes    []*model.Size
		Packages int
		Files    int
	}
	if err := json.Unmarshal([]byte(out.String()), &report); err != nil {
		t.Fatalf("json: %v\n%s", err, out.String())
	}
	if report.Files != 3 {
		t.Errorf("files = %d, want 3", report.Files)
	}
	if len(report.Sizes) == 0 {
		t.Error("the summary carries no histogram")
	}
	if strings.Contains(out.String(), "a.go") {
		t.Errorf("the summary lists the files:\n%s", out.String())
	}
}

// TestRenderJSON pins that the whole collection is written, files included.
func TestRenderJSON(t *testing.T) {
	var out strings.Builder
	if err := renderJSON(&out, testStats()); err != nil {
		t.Fatalf("renderJSON: %v", err)
	}
	var stats model.Stats
	if err := json.Unmarshal([]byte(out.String()), &stats); err != nil {
		t.Fatalf("json: %v", err)
	}
	if len(stats.Files) != 3 || len(stats.Histogram) == 0 {
		t.Errorf("collection = %d files, %d buckets", len(stats.Files), len(stats.Histogram))
	}
}

// TestParseFlags covers the renderers and the path argument, including the
// lone "-" the diagram pipeline writes for stdout.
func TestParseFlags(t *testing.T) {
	for _, tc := range []struct {
		name string
		args []string
		want options
	}{
		{name: "no arguments", args: nil, want: options{dir: "."}},
		{name: "d2 to stdout", args: []string{"-d2", "-"}, want: options{dir: ".", d2: true}},
		{name: "stats", args: []string{"-stats"}, want: options{dir: ".", stats: true}},
		{name: "a path", args: []string{"./cmd"}, want: options{dir: "./cmd"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, err := parseFlags(tc.args)
			if err != nil {
				t.Fatalf("parseFlags: %v", err)
			}
			if got.dir != tc.want.dir || got.d2 != tc.want.d2 || got.stats != tc.want.stats {
				t.Errorf("options = %+v, want %+v", *got, tc.want)
			}
		})
	}
}
