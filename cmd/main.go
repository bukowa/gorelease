package main

import (
	"github.com/bukforks/cobra"
	. "github.com/bukowa/gorelease"
	"log"
	"os"
)

// Globals
var (
	Version string
	Path    string
)

// GCSRelease
var (
	Bucket string
)

var rootCmd = &cobra.Command{
	Use:     "gorelease",
	Short:   "build and release your go application.",
	Version: Version,
}

var buildCmd = &cobra.Command{
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

var releaseCmd = &cobra.Command{
	Use:     "release",
	Short:   "release your targets",
	Version: Version,
}

var releaseGCS = &cobra.Command{
	Use:     "gcs",
	Short:   "release with google cloud storage",
	Version: Version,
	Run: func(cmd *cobra.Command, args []string) {
		var result = make(GCSResult)
		release := FromFile(Path)
		if err := Prepare(release); err != nil {
			log.Fatal(err)
		}
		if err := GCSRelease(Bucket, result)(release); err != nil {
			log.Fatal(err)
		}
		for k, v := range result {
			log.Print(k, "", v)
		}
	},
}

func main() {
	rootCmd.AddCommand(buildCmd)
	rootCmd.AddCommand(releaseCmd)
	releaseCmd.AddCommand(releaseGCS)

	buildCmd.Flags().StringVarP(&Path, "config", "c", ".gorelease.yaml", "path go gorelease config file")
	releaseGCS.Flags().StringVarP(&Bucket, "bucket", "b", "", "bucket name")

	if err := releaseGCS.MarkFlagRequired("bucket"); err != nil {
		log.Fatal(err)
	}

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
