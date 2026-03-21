package main

import (
	"fmt"
	"path"
	"strings"
)

// Rounded box-drawing characters
const (
	boxTopLeft     = "╭"
	boxTopRight    = "╮"
	boxBottomLeft  = "╰"
	boxBottomRight = "╯"
	boxHorizontal  = "─"
	boxVertical    = "│"
	boxTeeDown     = "┬"
	boxTeeUp       = "┴"
	boxTeeRight    = "├"
	boxTeeLeft     = "┤"
	boxCross       = "┼"
)

// cell holds both a plain string (for width calculation) and a colored string (for display).
type cell struct {
	plain   string
	colored string
}

func plainCell(s string) cell {
	return cell{plain: s, colored: s}
}

func coloredCell(plain, colored string) cell {
	return cell{plain: plain, colored: colored}
}

func emptyCell() cell {
	return cell{}
}

func shortName(modPath string) string {
	return path.Base(modPath)
}

func formatModuleCell(name string) cell {
	return coloredCell(name, colorBlue+name+colorReset)
}

func formatLatestCell(m moduleInfo) cell {
	if m.Latest == "" {
		return plainCell("")
	}
	if m.Ahead > 0 {
		ahead := fmt.Sprintf("%d commits ahead", m.Ahead)
		plain := fmt.Sprintf("%s (%s)", m.Latest, ahead)
		colored := fmt.Sprintf("%s%s%s %s(%s%s%s)%s", colorWhite, m.Latest, colorReset, colorGray, colorYellow, ahead, colorGray, colorReset)
		return coloredCell(plain, colored)
	}
	return coloredCell(m.Latest, colorWhite+m.Latest+colorReset)
}

func formatGitCell(m moduleInfo, st *gitStatus) cell {
	if st == nil {
		return plainCell("")
	}
	plain := m.Git
	var parts []string
	if st.Unpushed > 0 {
		parts = append(parts, fmt.Sprintf("%sunpushed: %d commits%s", colorRed, st.Unpushed, colorReset))
	}
	if st.Modified > 0 {
		parts = append(parts, fmt.Sprintf("%s%d modified%s", colorYellow, st.Modified, colorReset))
	}
	if st.Insertions > 0 || st.Deletions > 0 {
		parts = append(parts, fmt.Sprintf("%s+%d%s/%s-%d%s lines",
			colorGreen, st.Insertions, colorReset,
			colorOrange, st.Deletions, colorReset))
	}
	return coloredCell(plain, strings.Join(parts, colorGray+", "+colorReset))
}

func formatCommitMsgCell(msg string) cell {
	return coloredCell(msg, colorDim+msg+colorReset)
}

func formatDepListCell(paths []string) cell {
	if len(paths) == 0 {
		return plainCell("")
	}
	var plain, colored []string
	for _, p := range paths {
		name := shortName(p)
		plain = append(plain, name)
		colored = append(colored, colorWhite+name+colorReset)
	}
	return coloredCell(
		strings.Join(plain, ", "),
		strings.Join(colored, colorGray+", "+colorReset),
	)
}

// formatUsedByCell colors each dependent green if it uses the latest version of modPath, orange otherwise.
func formatUsedByCell(modPath string, usedByPaths []string, versionRefs map[string]map[string]string, latestTags map[string]string) cell {
	if len(usedByPaths) == 0 {
		return plainCell("")
	}
	latest := latestTags[modPath]
	var plain, colored []string
	for _, dep := range usedByPaths {
		name := shortName(dep)
		plain = append(plain, name)
		c := colorGreen
		if latest != "" {
			if refs, ok := versionRefs[dep]; ok {
				if ver, ok := refs[modPath]; ok && ver != latest {
					c = colorOrange
				}
			}
		}
		colored = append(colored, c+name+colorReset)
	}
	return coloredCell(
		strings.Join(plain, ", "),
		strings.Join(colored, colorGray+", "+colorReset),
	)
}

// tableRow represents one logical row which may span multiple display lines.
type tableRow struct {
	lines [][]cell // lines[lineIdx][colIdx]
}

func renderTables(modules []moduleInfo, versionRefs map[string]map[string]string, latestTags map[string]string, gitStatuses map[string]*gitStatus) {
	headers := []string{"Module", "Latest", "Git", "Used By", "Uses"}
	numCols := len(headers)

	var rows []tableRow
	for _, m := range modules {
		gitCell := formatGitCell(m, gitStatuses[m.Name])
		if len(m.GitMsgs) > 0 {
			gitCell = formatCommitMsgCell(m.GitMsgs[0])
		}
		first := []cell{
			formatModuleCell(m.Name),
			formatLatestCell(m),
			gitCell,
			formatUsedByCell(m.Name, m.UsedBy, versionRefs, latestTags),
			formatDepListCell(m.Uses),
		}
		tr := tableRow{lines: [][]cell{first}}

		// Add remaining commit message lines
		for _, msg := range m.GitMsgs[min(1, len(m.GitMsgs)):] {
			extra := make([]cell, numCols)
			for i := range extra {
				extra[i] = emptyCell()
			}
			extra[2] = formatCommitMsgCell(msg)
			tr.lines = append(tr.lines, extra)
		}

		rows = append(rows, tr)
	}

	// Compute column widths across all lines of all rows
	widths := make([]int, numCols)
	for i, h := range headers {
		widths[i] = len(h)
	}
	for _, tr := range rows {
		for _, line := range tr.lines {
			for i, c := range line {
				if len(c.plain) > widths[i] {
					widths[i] = len(c.plain)
				}
			}
		}
	}

	// Top border
	printBorder(boxTopLeft, boxTeeDown, boxTopRight, widths)

	// Header row
	printHeaderRow(headers, widths)

	// Header separator
	printBorder(boxTeeRight, boxCross, boxTeeLeft, widths)

	// Data rows
	for _, tr := range rows {
		for _, line := range tr.lines {
			printCellRow(line, widths)
		}
	}

	// Bottom border
	printBorder(boxBottomLeft, boxTeeUp, boxBottomRight, widths)

	// Count outdated dependencies
	outdated := 0
	for _, m := range modules {
		refs := versionRefs[m.Name]
		for _, dep := range m.Uses {
			latest := latestTags[dep]
			if latest != "" && refs != nil && refs[dep] != "" && refs[dep] != latest {
				outdated++
			}
		}
	}
	if outdated > 0 {
		fmt.Printf("%srun with %s-u%s %sto update %d outdated dependencies in workspace%s\n", colorGray, colorYellow, colorReset, colorGray, outdated, colorReset)
	}
}

func printBorder(left, mid, right string, widths []int) {
	var segs []string
	for _, w := range widths {
		segs = append(segs, strings.Repeat(boxHorizontal, w+2))
	}
	fmt.Println(colorGray + left + strings.Join(segs, mid) + right + colorReset)
}

func printHeaderRow(headers []string, widths []int) {
	var cells []string
	for i, h := range headers {
		cells = append(cells, fmt.Sprintf(" %s%s%-*s%s ", colorBold, colorMagenta, widths[i], h, colorReset))
	}
	fmt.Println(colorGray + boxVertical + colorReset +
		strings.Join(cells, colorGray+boxVertical+colorReset) +
		colorGray + boxVertical + colorReset)
}

func printCellRow(row []cell, widths []int) {
	var cells []string
	for i, c := range row {
		pad := widths[i] - len(c.plain)
		cells = append(cells, " "+c.colored+strings.Repeat(" ", pad)+" ")
	}
	fmt.Println(colorGray + boxVertical + colorReset +
		strings.Join(cells, colorGray+boxVertical+colorReset) +
		colorGray + boxVertical + colorReset)
}
