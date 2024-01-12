package internal

import (
	"fmt"
	"os"
	"path"
	"regexp"
	"strings"

	marecmd "github.com/femnad/mare/cmd"
)

const (
	goPlatform     = "linux-x86_64 "
	versionPattern = "(v)?[0-9]+\\.[0-9]+\\.[0-9]+"
)

type goApp struct {
	repo     string
	topLevel string
}

func (g goApp) assetDir() (string, error) {
	return g.topLevel, nil
}

func (g goApp) assetFile(version string) string {
	return fmt.Sprintf("%s-%s-%s", g.repo, version, goPlatform)
}

func (g goApp) canCompile() (bool, error) {
	return canCompileWith("go.mod", g.topLevel)
}

func (g goApp) cleanup() error {
	artifact := path.Join(g.topLevel, g.repo)
	_, err := os.Stat(artifact)
	if os.IsNotExist(err) {
		return nil
	} else if err != nil {
		return err
	}

	return os.Remove(artifact)
}

func (g goApp) compile() error {
	input := marecmd.Input{
		Command: "go build",
		Env:     map[string]string{"CGO_ENABLED": "0"},
	}
	_, err := marecmd.RunFmtErr(input)
	return err
}

func (g goApp) currentVersion() (string, error) {
	input := marecmd.Input{Command: "go run main.go --version"}
	out, err := marecmd.RunFmtErr(input)
	if err != nil {
		return "", err
	}

	fields := strings.Split(strings.TrimSpace(out.Stdout), " ")
	version := fields[len(fields)-1]

	re, err := regexp.Compile(versionPattern)
	if err != nil {
		return "", err
	}

	if !re.MatchString(version) {
		return "", fmt.Errorf("version %s does not match expected format", version)
	}

	return version, nil
}

func (goApp) name() string {
	return "Go"
}

func goCompiler(repo, topLevel string) compiler {
	return goApp{repo: repo, topLevel: topLevel}
}
