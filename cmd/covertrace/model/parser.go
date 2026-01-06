package model

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func ParseCoverageFile(filename string) ([]CoverageInfo, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	out, err := exec.Command("go", "list", ".").CombinedOutput()
	pkg := strings.TrimSpace(string(out))

	fmt.Fprintln(os.Stderr, "package:", pkg)

	var coverageData []CoverageInfo
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if !strings.HasPrefix(line, "mode:") {
			parts := strings.Fields(line)
			fileParts := strings.Split(parts[0], ":")
			filePath := fileParts[0]
			lines := strings.Split(fileParts[1], ",")
			startLine := parseLineNum(lines[0])
			endLine := parseLineNum(lines[1])
			numStmts := parseLineNum(parts[1])
			numCov := parseLineNum(parts[2])

			rawFile := filePath
			if strings.HasPrefix(filePath, pkg) {
				filePath = filePath[len(pkg)+1:]
				filePath = filepath.Join(".", filePath)
			}

			// Only include entries with coverage
			if numCov == 0 {
				continue
			}

			coverageData = append(coverageData, CoverageInfo{
				File:        filePath,
				RawFile:     rawFile,
				PackageName: pkg,
				StartLine:   startLine,
				EndLine:     endLine,
				NumStmts:    numStmts,
				NumCov:      numCov,
			})
		}
	}
	return coverageData, scanner.Err()
}

func ParseAllPackages(filename string) ([]string, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	out, err := exec.Command("go", "list", ".").CombinedOutput()
	pkg := strings.TrimSpace(string(out))

	packageSet := make(map[string]bool)
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if !strings.HasPrefix(line, "mode:") {
			parts := strings.Fields(line)
			fileParts := strings.Split(parts[0], ":")
			filePath := fileParts[0]

			// Extract subdirectory from raw file path
			dir := filepath.Dir(filePath)
			if strings.HasPrefix(dir, pkg) {
				subDir := dir[len(pkg):]
				subDir = strings.TrimPrefix(subDir, "/")
				if subDir != "" {
					packageSet[pkg+"/"+subDir] = true
				} else {
					packageSet[pkg] = true
				}
			}
		}
	}

	packages := make([]string, 0, len(packageSet))
	for p := range packageSet {
		packages = append(packages, p)
	}
	return packages, scanner.Err()
}

func parseLineNum(line string) int {
	var num int
	fmt.Sscanf(line, "%d.%d", &num, new(int))
	return num
}
