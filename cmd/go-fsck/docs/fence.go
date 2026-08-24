package docs

import "strings"

// fenceCodeBlocks turns the indented code blocks of a doc comment into
// fenced ones.
//
// A doc comment writes a code block by indenting it, which markdown renders
// as code without saying what language it is. In Go source it's Go, so each
// run of indented lines becomes a ```go block with one level of indentation
// removed.
func fenceCodeBlocks(doc string) string {
	var (
		out    []string
		block  []string
		blanks int
	)

	// flush writes the buffered block, keeping the blank lines which
	// followed it outside the fence.
	flush := func() {
		if len(block) == 0 {
			return
		}

		out = append(out, "```go")
		for _, line := range block {
			out = append(out, strings.TrimPrefix(line, "\t"))
		}
		out = append(out, "```")

		for range blanks {
			out = append(out, "")
		}

		block, blanks = nil, 0
	}

	for _, line := range strings.Split(doc, "\n") {
		switch {
		case strings.HasPrefix(line, "\t"):
			// A blank line between two indented ones is part of the block.
			for range blanks {
				block = append(block, "")
			}
			blanks = 0
			block = append(block, line)

		case strings.TrimSpace(line) == "" && len(block) > 0:
			blanks++

		default:
			flush()
			out = append(out, line)
		}
	}
	flush()

	return strings.Join(out, "\n")
}
