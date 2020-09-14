package cmd

import (
	"github.com/bukforks/cobra"
	. "github.com/bukowa/gorelease"
	"log"
)

var BuildCmd = &cobra.Command{
	Use:     "build",
	Short:   "go build targets",
	Version: Version,
	Run: func(cmd *cobra.Command, args []string) {
		release := FromFile(Path)

		// prepare
		if err := Prepare(release); err != nil {
			log.Fatal(err)
		}

		// build
		if err := BuildRelease(release); err != nil {
			log.Fatal(err)
		}
	},
}

func init() {
	BuildCmd.Flags().StringVarP(&Path, "config", "c", ".gorelease.yaml", "path go gorelease config file")
}
