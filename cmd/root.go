package cmd

import (
	"github.com/bukforks/cobra"
)

// Globals
var (
	Version string
	Path    string
)

var RootCmd = &cobra.Command{
	Use:     "gorelease",
	Short:   "build and release your go application.",
	Version: Version,
}

func init() {
	RootCmd.AddCommand(BuildCmd)
	RootCmd.AddCommand(ReleaseCmd)

}
