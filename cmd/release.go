package cmd

import "github.com/bukforks/cobra"

var ReleaseCmd = &cobra.Command{
	Use:     "release",
	Short:   "release your targets",
	Version: Version,
}

func init() {
	ReleaseCmd.AddCommand(ReleaseGCS)
}
