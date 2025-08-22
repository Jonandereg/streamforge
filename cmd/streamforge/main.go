package main

import (
	"fmt"
	"os"

	"github.com/jonandereg/streamforge/internal/version"
)

func main() {
	fmt.Printf("StreamForge %s (commit %s, built %s)\n", version.Version, version.Commit, version.BuildDate)
	os.Exit(0)
}
