package gorelease

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
	"path"
	"strings"
)

type (
	BuildFunc        func(*Target) error
	BuildReleaseFunc func(*Release) error
	PrepareFunc      func(*Release) error
	ReleaseFunc      func(*Release) error
)

type Release struct {
	// todo raw Target instead of Global
	Global  Target   `yaml:"global"`
	Targets []Target `yaml:"targets"`
}

type Target struct {
	FilePath string `yaml:"file"`    // filename to compile
	NameFmt  string `yaml:"name"`    // name pattern to apply to compiled file
	Version  string `yaml:"version"` // version of release
	DestDir  string `yaml:"dir"`     // dir when bin is stored

	Env       map[string]string   `yaml:"env"`       // env passed to go build
	Flags     []string            `yaml:"flags"`     // flags passed to go build
	Platforms map[string][]string `yaml:"platforms"` // for what platforms build

	FileBuilds []FileBuild `yaml:"-"`
}

type FileBuild struct {
	Name    string   // file name
	BinPath string   // final path of built executable
	Env     []string // env passed to go build
	Args    []string // args passed to go build
}

func (b *FileBuild) Command() *exec.Cmd {
	cmd := exec.Command("go", b.Args...)
	cmd.Env = b.Env
	return cmd
}

// ForEach performs func f for each Target in Release
func (r *Release) ForEachTarget(f func(t *Target) error) error {
	for _, target := range r.Targets {
		if err := f(&target); err != nil {
			return err
		}
	}
	return nil
}

func (r *Release) ForEachTargetBuild(f func(t *Target, f *FileBuild) error) error {
	for _, target := range r.Targets {
		for _, build := range target.FileBuilds {
			err := f(&target, &build)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// Build is a basic BuildFunc
var Build BuildFunc = func(target *Target) error {
	for _, b := range target.FileBuilds {
		cmd := b.Command()
		log.Print(logCmdString(cmd))
		out, _, err := runCmd(cmd)
		if len(out) > 0 {
			log.Print(string(out))
		}
		if err != nil {
			return err
		}
	}
	return nil
}

var BuildRelease = func(release *Release) error {
	for _, t := range release.Targets {
		if err := Build(&t); err != nil {
			return err
		}
	}
	return nil
}

// Prepare is a basic PrepareFunc
var Prepare PrepareFunc = func(release *Release) error {

	// check if release is for all platforms
	if isAllPlatforms(release.Global.Platforms) {
		release.Global.Platforms = DistList()
	}
	// shorten var name
	glob := release.Global

	// iterate over each target
	for i, t := range release.Targets {

		if t.Version == "" {
			t.Version = glob.Version
		}
		if t.DestDir == "" {
			t.DestDir = glob.DestDir
		}
		if t.FilePath == "" {
			return ErrorEmptyFileName(t)
		}
		if t.NameFmt == "" {
			t.NameFmt = glob.NameFmt
		}
		if t.Env == nil {
			t.Env = glob.Env
		}
		if t.Flags == nil {
			t.Flags = glob.Flags
		}
		if isAllPlatforms(t.Platforms) {
			t.Platforms = DistList()
		}
		if t.Platforms == nil {
			t.Platforms = glob.Platforms
		}

		for goos, goarchs := range t.Platforms {
			for _, goarch := range goarchs {
				build := MakeFileBuild(t, goos, goarch)
				t.FileBuilds = append(t.FileBuilds, build)
			}
		}
		release.Targets[i] = t
	}
	return nil
}

// MakeFileBuild creates FileBuild for Target
func MakeFileBuild(t Target, goos, goarch string) FileBuild {
	name := fmtName(t.NameFmt, t.Version, goos, goarch)
	bin := path.Join(t.DestDir, name)
	flags := append(t.Flags, "-o", bin)
	env := buildEnv(t, goos, goarch)
	args := append(append([]string{"build"}, flags...), t.FilePath)
	b := FileBuild{
		Name:    name,
		BinPath: bin,
		Env:     env,
		Args:    args,
	}
	return b
}

// DistList is a function that will gather all dists
// by calling `go tool dist list`
func DistList() map[string][]string {
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

type ErrorEmptyFileName Target

func (err ErrorEmptyFileName) Error() string {
	return fmt.Sprintf("target '%#v' has empty file name", err)
}

// FromFile creates Release from given path
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

func cmdErr(cmd *exec.Cmd, output []byte, code int, err error) error {
	return errors.Wrapf(err,
		"%s - error while executing command: '%s'"+
			" - in directory: '%s' - exit code: '%v' - error: '%s'",
		output, cmd.String(), cmdWd(cmd), code, err)
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

func isAllPlatforms(p map[string][]string) bool {
	if _, ok := p["all"]; ok {
		return true
	}
	return false
}

func envMapToSlice(m map[string]string) (env []string) {
	for k, v := range m {
		env = append(env, fmt.Sprintf("%s=%s", k, v))
	}
	return env
}

func appendEnvMap(env []string, m map[string]string) []string {
	return append(env, envMapToSlice(m)...)
}

func buildEnv(t Target, goos, goarch string) []string {
	env := os.Environ()
	env = appendEnvMap(env, t.Env)
	env = appendEnvKeyValue(env, "GOOS", goos)
	env = appendEnvKeyValue(env, "GOARCH", goarch)
	return env
}

func appendEnvKeyValue(env []string, k, v string) []string {
	return append(env, fmt.Sprintf("%s=%s", k, v))
}

func fmtName(name, version, goos, goarch string) string {
	return fmt.Sprintf(name, version, goos, goarch)
}

func logCmdString(cmd *exec.Cmd) string {
	return fmt.Sprintf("executing '%s' in '%s'", cmd.String(), cmdWd(cmd))
}

func logBuild(t Target, cmd *exec.Cmd) {
	log.Printf("%s - for target: %s", logCmdString(cmd), t)
}
