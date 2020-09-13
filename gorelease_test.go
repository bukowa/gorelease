package main

import (
	"log"
	"reflect"
	"runtime"
	"testing"
)

func Test_loadConfig(t *testing.T) {
	cfg := loadYaml("gorelease.yaml")
	log.Print(cfg)
}

func Test_goToolDistList(t *testing.T) {
	dist := goToolDistList()
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
			Name: "rel%s-%s",
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
				File: "main.go",
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
				File: "main.go",
				Name: "new",
			},
		},
	}
	err := NewHandler().Prepare(r)
	if err != nil {
		t.Error(err)
	}

	if !reflect.DeepEqual(r.Global.Platforms, goToolDistList()) {
		t.Errorf("got: %s, want: %s", r.Global, goToolDistList())
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

	if target.Name != "new" {
		t.Error(target.Name)
	}
	if !reflect.DeepEqual(target.Platforms, goToolDistList()) {
		t.Errorf("got: %s, want: %s", target.Platforms, goToolDistList())
	}
}

func TestTarget_Build(t *testing.T) {
	r := getExampleRelease()
	h := NewHandler()
	if err := h.Prepare(r); err != nil {
		t.Error(err)
	}

	target := r.Targets[0]
	// todo test
	target.Command(runtime.GOOS, runtime.GOARCH)
}

func Test_handler_Build(t *testing.T) {
	r := getExampleRelease()
	h := NewHandler()
	if err := h.Prepare(r); err != nil {
		t.Error(err)
	}

	for _, target := range r.Targets {
		if err := h.Build(&target); err != nil {
			t.Error(err)
		}
	}
}

func getExampleRelease() *Release {
	return FromFile("./examples/gorelease.yaml")
}
