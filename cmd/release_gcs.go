package cmd

import (
	"github.com/bukforks/cobra"
	. "github.com/bukowa/gorelease"
	"log"
)

var Bucket string

var ReleaseGCS = &cobra.Command{
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
			log.Print(k, " ", v)
		}
	},
}

func init() {
	ReleaseGCS.Flags().StringVarP(&Bucket, "bucket", "b", "", "bucket name")
	if err := ReleaseGCS.MarkFlagRequired("bucket"); err != nil {
		log.Fatal(err)
	}
}
