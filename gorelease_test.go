package gorelease_test

import (
	"fmt"
	. "github.com/bukowa/gorelease"
	"log"
	"os"
	"path"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"testing"
)

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

func Test_handler_Prepare(t *testing.T) {
	r := &Release{
		Global: Target{
			NameFmt: "rel%s-%s",
			Env: map[string]string{
				"key": "value",
			},
			Platforms: map[string][]string{
				"all": nil,
			},
			Flags: nil,
		},
		Targets: []Target{
			{
				FilePath: "main.go",
				Env: map[string]string{
					"CGO_ENABLED": "0",
					"GO111MODULE": "on",
				},
				Platforms: map[string][]string{
					"windows": {"386", "amd64"},
				},
				Flags: nil,
			},
			{
				FilePath: "main.go",
				NameFmt:  "new",
			},
		},
	}
	err := Prepare(r)
	if err != nil {
		t.Error(err)
	}

	if !reflect.DeepEqual(r.Global.Platforms, DistList()) {
		t.Errorf("got: %s, want: %s", r.Global, DistList())
	}

	// target0
	target := r.Targets[0]

	if !reflect.DeepEqual(target.Platforms, map[string][]string{"windows": {"386", "amd64"}}) {
		t.Error(target.Platforms)
	}
	if !reflect.DeepEqual(target.Env, map[string]string{"CGO_ENABLED": "0", "GO111MODULE": "on"}) {
		t.Error(target.Env)
	}

	// target1
	target = r.Targets[1]

	if target.NameFmt != "new" {
		t.Error(target.NameFmt)
	}
	if !reflect.DeepEqual(target.Platforms, DistList()) {
		t.Errorf("got: %s, want: %s", target.Platforms, DistList())
	}
}

func Test_handler_Build(t *testing.T) {
	os.RemoveAll("./examples/bin")

	r := getExampleRelease()
	if err := Prepare(r); err != nil {
		t.Error(err)
	}

	for _, target := range r.Targets {
		if err := Build(&target); err != nil {
			t.Error(err)
		}
	}

	// count number of built files
	var wantFiles = totalFiles(r)
	var gotFiles []os.FileInfo
	err := filepath.Walk("./examples/bin", func(path string, info os.FileInfo, err error) error {
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

func getExampleRelease() *Release {
	return FromFile("./examples/.gorelease.yaml")
}

func totalFiles(r *Release) (n int) {
	for _, t := range r.Targets {
		for _, a := range t.Platforms {
			n += len(a)
		}
	}
	return
}

func Test_fileBuild(t *testing.T) {
	type args struct {
		t      Target
		goos   string
		goarch string
	}
	tests := []struct {
		name string
		args args
		want FileBuild
	}{
		{
			name: "1",
			args: args{
				t: Target{
					Version:  "0.1.0",
					DestDir:  "bin",
					FilePath: "main.go",
					NameFmt:  "test%s-%s-%s",
					Env: map[string]string{
						"key": "value",
					},
					Platforms:  map[string][]string{},
					Flags:      []string{},
					FileBuilds: []FileBuild{},
				},
				goos:   "linux",
				goarch: "amd64",
			},
			want: FileBuild{
				Name:    "test0.1.0-linux-amd64",
				BinPath: path.Join("bin", "test0.1.0-linux-amd64"),
				Env: []string{
					"key=value",
				},
				Args: []string{
					"build", "-o", path.Join("bin", "test0.1.0-linux-amd64"), "main.go",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := MakeFileBuild(tt.args.t, tt.args.goos, tt.args.goarch)

			// check env
			if !envMapInSlice(tt.args.t.Env, got.Env) {
				t.Error(got.Env)
			}

			// cmd check
			cmd := got.Command()
			if strings.Join(cmd.Args, " ") != strings.Join(append([]string{"go"}, tt.want.Args...), " ") {
				t.Errorf("got: %s, want: %s", cmd.Args, append([]string{"go"}, tt.want.Args...))
			}

			if !envMapInSlice(tt.args.t.Env, cmd.Env) {
				t.Error()
			}

			// clear env b-c it contains os env
			got.Env = []string{}
			tt.want.Env = []string{}

			// reflect check
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("MakeFileBuild() = %v, want %v", got, tt.want)
			}

		})
	}
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
