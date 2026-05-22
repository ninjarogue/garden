package main

import (
	"fmt"
	"os"

	"github.com/aric/garden/internal/app"
	gardencmd "github.com/aric/garden/internal/cmd"
)

func main() {
	garden := app.New(app.Options{Root: "."})
	root := gardencmd.NewRoot(gardencmd.Options{App: garden, Stdout: os.Stdout, Stderr: os.Stderr})
	if err := root.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
