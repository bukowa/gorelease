package gorelease_test

import . "github.com/bukowa/gorelease"

import (
	"os"
	"testing"
)

var (
	Bucket = bucketName()
)

func TestGCSReleaser(t *testing.T) {
	os.RemoveAll("bin")
	r := prepareExample()
	err := BuildRelease(r)
	if err != nil {
		t.Error(err)
	}
	err = GCSRelease(Bucket)(r)
	if err != nil {
		t.Error(err)
	}
}

func bucketName() string {
	if b := os.Getenv("TEST_RELEASE_BUCKET"); b == "" {
		panic("TEST_RELEASE_BUCKET")
	} else {
		return b
	}
}
