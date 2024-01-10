package internal

import (
	"fmt"
	"os"
	"strings"

	marecmd "github.com/femnad/mare/cmd"
)

const (
	goPlatform = "linux-x86_64 "
)

type goApp struct {
	executable string
	topLevel   string
}

func (g goApp) assetDir() (string, error) {
	return g.topLevel, nil
}

func (g goApp) assetFile(executable, version string) string {
	g.executable = executable
	return fmt.Sprintf("%s-%s-%s", executable, version, goPlatform)
}

func (g goApp) canCompile() (bool, error) {
	return canCompileWith("go.mod", g.topLevel)
}

func (g goApp) cleanup() error {
	if g.executable == "" {
		return nil
	}

	return os.Remove(g.executable)
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
	return fields[len(fields)-1], nil
}

func goCompiler(topLevel string) compiler {
	return goApp{topLevel: topLevel}
}
