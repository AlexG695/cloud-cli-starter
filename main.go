package main

import (
	"fmt"
	"os"

	"github.com/tu-usuario/cloud/cmd"
)

func main() {
	if err := cmd.InitCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error executing CLI: %v\n", err)
		os.Exit(1)
	}
}
