package main

import (
	"bufio"
	"bytes"
	"fmt"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	ppath "path"
	"strings"
)

type Handler interface {
	Prepare(release *Release) error
	Build(target *Target) error
	//Release(target *Target) error
}

type Release struct {
	Global  Target   `yaml:"global"`
	Targets []Target `yaml:"targets"`
}

type Target struct {
	Workdir   string              `yaml:"workdir"`   // where go build is run
	Version   string              `yaml:"version"`   // version of release
	Dir       string              `yaml:"dir"`       // dir when bin is stored
	File      string              `yaml:"file"`      // filename to compile
	Name      string              `yaml:"name"`      // name pattern to apply to compiled file
	Env       map[string]string   `yaml:"env"`       // env passed to go build
	Platforms map[string][]string `yaml:"platforms"` // for what platforms build
	Flags     []string            `yaml:"flags"`     // flags passed to go build
	BinPath   string              `yaml:"-"`         // final path of built executable
}

func (t *Target) Command(goos, goarch string) *exec.Cmd {
	file := fmtName(t.Name, t.Version, goos, goarch)
	path := ppath.Join(t.Dir, file)

	t.Flags = append(t.Flags, "-o", path)
	cmd := exec.Command("go", append([]string{"build"}, t.Flags...)...)
	appendMapEnv(cmd, t.Env)
	setKVEnv(cmd, "GOOS", goos)
	setKVEnv(cmd, "GOARCH", goarch)
	cmd.Args = append(cmd.Args, t.File)

	if t.Workdir != "" {
		cmd.Dir = t.Workdir
	}
	t.BinPath = ppath.Join(t.Workdir, path)
	return cmd
}

// handler implements Handler
type handler struct{}

func (h *handler) Build(t *Target) error {
	for goos, goarchs := range t.Platforms {
		for _, arch := range goarchs {
			cmd := t.Command(goos, arch)
			logBuild(*t, cmd)
			out, _, err := runCmd(cmd)
			if err != nil {
				return errors.Wrap(err, string(out))
			}
			log.Printf("successfully built %s", t.BinPath)
		}
	}
	return nil
}

func (h *handler) Prepare(r *Release) error {

	// check if release is for all platforms
	if isAllPlatforms(r.Global.Platforms) {
		r.Global.Platforms = goToolDistList()
	}
	// shorten var name
	glob := r.Global

	// iterate over each target
	for i, t := range r.Targets {
		if t.Workdir == "" {
			t.Workdir = glob.Workdir
		}
		if t.Version == "" {
			t.Version = glob.Version
		}
		if t.Dir == "" {
			t.Dir = glob.Dir
		}
		if t.File == "" {
			return ErrorEmptyFileName(t)
		}
		if t.Name == "" {
			t.Name = glob.Name
		}
		if t.Env == nil {
			t.Env = glob.Env
		}
		if t.Flags == nil {
			t.Flags = glob.Flags
		}
		if isAllPlatforms(t.Platforms) {
			t.Platforms = goToolDistList()
		}
		if t.Platforms == nil {
			t.Platforms = glob.Platforms
		}
		r.Targets[i] = t
	}
	return nil
}

func (h *handler) Release(target Target) error {
	panic("implement me")
}

func NewHandler() Handler {
	return &handler{}
}

func FromFile(path string) *Release {
	return loadYaml(path)
}

func loadYaml(path string) (cfg *Release) {
	f, err := os.Open(path)
	if err != nil {
		log.Fatal(err)
	}
	b, err := ioutil.ReadAll(f)
	if err != nil {
		log.Fatal(err)
	}
	err = yaml.Unmarshal(b, &cfg)
	if err != nil {
		log.Fatal(errors.Wrap(err, path))
	}
	return
}

func isAllPlatforms(p map[string][]string) bool {
	if _, ok := p["all"]; ok {
		return true
	}
	return false
}

func isReqAllPlatforms(r *Release) bool {
	if isAllPlatforms(r.Global.Platforms) {
		return true
	}
	for _, t := range r.Targets {
		if isAllPlatforms(t.Platforms) {
			return true
		}
	}
	return false
}

func goToolDistList() map[string][]string {
	cmd := exec.Command("go", "tool", "dist", "list")
	output := runCmdFatal(cmd)
	var m = make(map[string][]string)
	scanner := bufio.NewScanner(bytes.NewReader(output))
	for scanner.Scan() {
		s := strings.Split(scanner.Text(), "/")
		m[s[0]] = append(m[s[0]], s[1])
	}
	return m
}

func runCmd(cmd *exec.Cmd) (output []byte, code int, err error) {
	stdout, stderr := stdouts(cmd)
	if err = cmd.Run(); err != nil {
		if v, ok := err.(*exec.ExitError); ok {
			code = v.ExitCode()
		}
	}
	output = append(output, read(stdout)...)
	output = append(output, read(stderr)...)
	return
}

func runCmdFatal(cmd *exec.Cmd) []byte {
	output, code, err := runCmd(cmd)
	if err != nil || code != 0 {
		log.Fatal(cmdErr(cmd, output, code, err))
	}
	return output
}

func cmdErr(cmd *exec.Cmd, output []byte, code int, err error) error {
	return errors.Wrapf(err, "%s - error while executing command: '%s' - in directory: '%s' - exit code: %v - error: %s", output, cmd.String(), cmdWd(cmd), code, err)
}

func stdouts(cmd *exec.Cmd) (stdout, stderr *bytes.Buffer) {
	stdout = bytes.NewBuffer(nil)
	stderr = bytes.NewBuffer(nil)
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	return stdout, stderr
}

func read(r io.Reader) []byte {
	b, err := ioutil.ReadAll(r)
	if err != nil {
		log.Fatal(err)
	}
	return b
}

func cmdWd(cmd *exec.Cmd) string {
	if cmd.Dir != "" {
		return cmd.Dir
	}
	wd, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	return wd
}

func convertEnv(m map[string]string) (env []string) {
	for k, v := range m {
		env = append(env, fmt.Sprintf("%s=%s", k, v))
	}
	return env
}

func appendMapEnv(cmd *exec.Cmd, m map[string]string) {
	cmd.Env = os.Environ()
	cmd.Env = append(cmd.Env, convertEnv(m)...)
}

func setKVEnv(cmd *exec.Cmd, k, v string) {
	cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", k, v))
}

func fmtName(name, version, goos, goarch string) string {
	return fmt.Sprintf(name, version, goos, goarch)
}

func logCmd(cmd *exec.Cmd) {
	log.Print(logCmdString(cmd))
}

func logCmdString(cmd *exec.Cmd) string {
	return fmt.Sprintf("executing '%s' in '%s'", cmd.String(), cmdWd(cmd))
}

func logBuild(t Target, cmd *exec.Cmd) {
	log.Printf("%s - for target: %s", logCmdString(cmd), t)
}
