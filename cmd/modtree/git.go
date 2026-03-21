package main

import (
	"bufio"
	"fmt"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

func getGitStatus(dir string) *gitStatus {
	absDir, err := filepath.Abs(dir)
	if err != nil {
		return nil
	}

	// Find git root to determine relative path for scoping
	cmd := exec.Command("git", "rev-parse", "--show-toplevel")
	cmd.Dir = absDir
	rootOut, err := cmd.Output()
	if err != nil {
		return nil
	}
	gitRoot := strings.TrimSpace(string(rootOut))

	// Relative path from git root to module dir (for scoping)
	relPath, err := filepath.Rel(gitRoot, absDir)
	if err != nil {
		return nil
	}
	isSubdir := relPath != "."

	st := &gitStatus{}

	// Count modified files (working tree + staged)
	args := []string{"status", "--porcelain"}
	if isSubdir {
		args = append(args, "--", relPath)
	}
	cmd = exec.Command("git", args...)
	cmd.Dir = gitRoot
	out, err := cmd.Output()
	if err == nil {
		for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
			if line != "" {
				st.Modified++
			}
		}
	}

	// Get diff stats (insertions/deletions)
	args = []string{"diff", "--shortstat"}
	if isSubdir {
		args = append(args, "--", relPath)
	}
	cmd = exec.Command("git", args...)
	cmd.Dir = gitRoot
	out, err = cmd.Output()
	if err == nil {
		parseShortstat(strings.TrimSpace(string(out)), st)
	}

	// Also include staged changes in the counts
	args = []string{"diff", "--cached", "--shortstat"}
	if isSubdir {
		args = append(args, "--", relPath)
	}
	cmd = exec.Command("git", args...)
	cmd.Dir = gitRoot
	out, err = cmd.Output()
	if err == nil {
		var staged gitStatus
		parseShortstat(strings.TrimSpace(string(out)), &staged)
		st.Insertions += staged.Insertions
		st.Deletions += staged.Deletions
	}

	// Count unpushed commits (scoped to subtree if applicable)
	args = []string{"log", "--oneline", "@{u}..HEAD"}
	if isSubdir {
		args = append(args, "--", relPath)
	}
	cmd = exec.Command("git", args...)
	cmd.Dir = gitRoot
	out, err = cmd.Output()
	if err == nil {
		for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
			if line != "" {
				st.Unpushed++
			}
		}
	}

	if st.Unpushed == 0 && st.Modified == 0 && st.Insertions == 0 && st.Deletions == 0 {
		return nil
	}
	return st
}

// parseShortstat parses "N files changed, N insertions(+), N deletions(-)"
func parseShortstat(s string, st *gitStatus) {
	for _, part := range strings.Split(s, ",") {
		part = strings.TrimSpace(part)
		fields := strings.Fields(part)
		if len(fields) < 2 {
			continue
		}
		n, err := strconv.Atoi(fields[0])
		if err != nil {
			continue
		}
		switch {
		case strings.Contains(part, "insertion"):
			st.Insertions += n
		case strings.Contains(part, "deletion"):
			st.Deletions += n
		}
	}
}

func latestGitTag(dir string) string {
	cmd := exec.Command("git", "tag", "--list", "--sort=-v:refname", "v*")
	cmd.Dir = dir
	out, err := cmd.Output()
	if err != nil {
		return ""
	}
	scanner := bufio.NewScanner(strings.NewReader(string(out)))
	if scanner.Scan() {
		return strings.TrimSpace(scanner.Text())
	}
	return ""
}

func commitsSinceTag(dir, tag string) int {
	cmd := exec.Command("git", "rev-list", "--count", tag+"..HEAD", "--", ".")
	cmd.Dir = dir
	out, err := cmd.Output()
	if err != nil {
		return 0
	}
	n, _ := strconv.Atoi(strings.TrimSpace(string(out)))
	return n
}

func commitMessagesSinceTag(dir, tag string) []string {
	cmd := exec.Command("git", "log", "--oneline", "--format=%s", tag+"..HEAD", "--", ".")
	cmd.Dir = dir
	out, err := cmd.Output()
	if err != nil {
		return nil
	}
	var msgs []string
	for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		if line != "" {
			msgs = append(msgs, line)
		}
	}
	return msgs
}

func formatGitSummary(st *gitStatus) string {
	var parts []string
	if st.Unpushed > 0 {
		parts = append(parts, fmt.Sprintf("unpushed: %d commits", st.Unpushed))
	}
	if st.Modified > 0 {
		parts = append(parts, fmt.Sprintf("%d modified", st.Modified))
	}
	if st.Insertions > 0 || st.Deletions > 0 {
		parts = append(parts, fmt.Sprintf("+%d/-%d lines", st.Insertions, st.Deletions))
	}
	return strings.Join(parts, ", ")
}
