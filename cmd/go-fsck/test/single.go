package test

import (
	"os"
	"os/exec"
)

// runSingleTest runs go test without special multi-module handling.
// This is used when -c is not passed.
func runSingleTest(args []string) error {
	cmd := exec.Command("go", append([]string{"test"}, args...)...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	return cmd.Run()
}
