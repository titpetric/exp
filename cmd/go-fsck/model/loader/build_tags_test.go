package loader_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/titpetric/exp/cmd/go-fsck/model/loader"
)

func TestBuildTags(t *testing.T) {
	src := `
// +build debug
// +build linux

package main

import "fmt"

func main() {
	fmt.Println("Hello, World!")
}`

	want := []string{"debug", "linux"}
	got := loader.BuildTags([]byte(src))

	assert.Equal(t, want, got)
}
