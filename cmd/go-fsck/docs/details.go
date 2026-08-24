package docs

import (
	"fmt"
	"strings"

	"github.com/titpetric/exp/cmd/go-fsck/model"
)

// details renders a type, const or var as a section a reader opens.
//
// A reference of any size is read by looking for one symbol, and github
// collapses <details> to its <summary>, so the page is the list of what a
// package declares with each declaration one click away. The blank lines
// around the block are what makes github render the markdown inside it.
func details(decl *model.Declaration, source string) string {
	var out strings.Builder

	fmt.Fprintf(&out, "<details>\n<summary><code>%s</code></summary>\n\n", declarationSummary(decl))
	fmt.Fprintf(&out, "```go\n%s\n```\n\n", source)
	fmt.Fprint(&out, "</details>\n\n")

	return out.String()
}
