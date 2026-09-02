package main

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/titpetric/exp/cmd/go-ddd-stats/model"
)

// renderJSON writes the whole collection: every file, every package and the
// histogram over them.
func renderJSON(w io.Writer, stats *model.Stats) error {
	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", "  ")
	return enc.Encode(stats)
}

// renderStats writes the histogram alone, for a reader who wants the shape of
// the tree rather than the files in it.
func renderStats(w io.Writer, stats *model.Stats) error {
	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", "  ")
	return enc.Encode(struct {
		Sizes    []*model.Size
		Packages int
		Files    int
	}{
		Sizes:    stats.Histogram,
		Packages: len(stats.Packages),
		Files:    len(stats.Files),
	})
}

// renderD2 writes the histogram as a d2 source document, one node per bucket
// holding the count of files in it. Empty buckets are left out: a bucket with
// no files is a gap in the range and not a measurement.
//
// The output goes through the d2 binary, which is what turns it into the
// diagram a README embeds:
//
//	go-ddd-stats -d2 | d2 --layout elk - docs/assets/size.svg
func renderD2(w io.Writer, stats *model.Stats) error {
	title := fmt.Sprintf("File size distribution (*.go, %d files)", len(stats.Files))
	if _, err := fmt.Fprintf(w, "title: |md\n  # %s\n| {near: top-center}\n\n", title); err != nil {
		return err
	}

	var previous string
	for i, bucket := range stats.Histogram {
		if bucket.Count == 0 {
			continue
		}
		id := fmt.Sprintf("b%d", i)
		label := strings.ReplaceAll(bucket.Size, `"`, "")
		if _, err := fmt.Fprintf(w, "%s: %q {\n  shape: rectangle\n  count: %d\n}\n", id, label, bucket.Count); err != nil {
			return err
		}
		if previous != "" {
			if _, err := fmt.Fprintf(w, "%s -> %s\n", previous, id); err != nil {
				return err
			}
		}
		previous = id
	}
	return nil
}
