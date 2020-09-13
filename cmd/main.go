package main

import (
	"github.com/bukforks/cobra"
	. "github.com/bukowa/gorelease"
	"log"
	"os"
)

var (
	Version string
	Path    string
)

var Command = &cobra.Command{
	Use:     "gorelease",
	Short:   "build and release your go application.",
	Version: Version,
}

var Build = &cobra.Command{
	Use:     "build",
	Short:   "go build targets",
	Version: Version,
	Run: func(cmd *cobra.Command, args []string) {
		release := FromFile(Path)
		handler := NewHandler()
		if err := handler.Prepare(release); err != nil {
			log.Fatal(err)
		}
		for _, t := range release.Targets {
			if err := handler.Build(&t); err != nil {
				log.Fatal(err)
			}
		}
	},
}

func main() {
	Command.AddCommand(Build)
	Build.Flags().StringVarP(&Path, "config", "c", ".gorelease.yaml", "path go gorelease config file")
	if err := Command.Execute(); err != nil {
		os.Exit(1)
	}
}
