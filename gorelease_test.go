package gorelease_test

import (
	"fmt"
	. "github.com/bukowa/gorelease"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestMain(m *testing.M) {
	if err := os.Chdir("./examples"); err != nil {
		log.Fatal(err)
	}
	os.Exit(m.Run())
}

func Test_goToolDistList(t *testing.T) {
	dist := DistList()
	if len(dist) < 1 {
		t.Error(dist)
	}
	log.Print(dist)
	for k, v := range dist {
		if k == runtime.GOOS {
			for _, vv := range v {
				if vv == runtime.GOARCH {
					return
				}
			}
		}
	}
	t.Error("not working")
}

func Test_handler_Build(t *testing.T) {
	os.RemoveAll("bin")
	r := prepareExample()

	for _, target := range r.Targets {
		if err := Build(&target); err != nil {
			t.Error(err)
		}
	}

	// count number of built files
	var wantFiles = totalFiles(r)
	var gotFiles []os.FileInfo
	err := filepath.Walk("bin", func(path string, info os.FileInfo, err error) error {
		if info != nil && !info.IsDir() {
			gotFiles = append(gotFiles, info)
		}
		return err
	})
	if err != nil {
		t.Error(err)
	}
	if len(gotFiles) != wantFiles {
		t.Errorf("got: %v want: %v", len(gotFiles), wantFiles)
	}
}

func prepareExample() *Release {
	release := exampleRelease()
	if err := Prepare(release); err != nil {
		log.Fatal(err)
	}
	return release
}

func envMapInSlice(m map[string]string, env []string) bool {
	for k, v := range m {
		for _, e := range env {
			if e == fmt.Sprintf("%s=%s", k, v) {
				return true
			}
		}
	}
	return false
}

func exampleRelease() *Release {
	return FromFile(".gorelease.yaml")
}

func totalFiles(r *Release) (n int) {
	for _, t := range r.Targets {
		for _, a := range t.Platforms {
			n += len(a)
		}
	}
	return
}
