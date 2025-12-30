package test

import (
	"fmt"
	"os"
	"path"
)

func PrintHelp() {
	fmt.Printf("Usage: %s test <go test args>\n\n", path.Base(os.Args[0]))
	fmt.Println("Description:")
	fmt.Println("  Wraps 'go test' with multi-module workspace support.")
	fmt.Println("  When -c is used with -o, compiles test binaries for all packages in the workspace")
	fmt.Println("  to the specified directory, handling package name conflicts.")
	fmt.Println("\nExamples:")
	fmt.Println("  go-fsck test ./...")
	fmt.Println("  go-fsck test -c -o bin ./...")
	fmt.Println("  go-fsck test -trimpath -tags=integration ./...")
}
