package main

import (
	"fmt"
	"os"

	"github.com/babarot/afx/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "[ERROR]: %v\n", err)
		os.Exit(1)
	}
}
