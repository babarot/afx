package main

import (
	"fmt"
	"os"

	"github.com/b4b4r07/afx/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "[ERROR]: %v\n", err)
		os.Exit(1)
	}
}
