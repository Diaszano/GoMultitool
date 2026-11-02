// Package main provides the multitool command-line interface.
package main

import (
	"fmt"
	"os"

	"github.com/Diaszano/GoMultitool/internal/buildinfo"
)

func main() {
	if len(os.Args) != 2 || os.Args[1] != "version" {
		fmt.Fprintln(os.Stderr, "Usage: multitool version")
		os.Exit(2)
	}

	info := buildinfo.Current()
	fmt.Printf("RK322x Multitool %s\nCommit: %s\nBuild: %s\n", info.Version, info.Commit, info.BuildTime)
}
