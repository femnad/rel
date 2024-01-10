package internal

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"

	marecmd "github.com/femnad/mare/cmd"
)

const (
	cargoConfig = "Cargo.toml"
	platform    = "x86_64-unknown-linux-gnu"
)

type rust struct {
	topLevel string
}

func (r rust) assetFile(executable, version string) string {
	return fmt.Sprintf("%s-%s-%s", executable, version, platform)
}

func (r rust) assetDir() (string, error) {
	err := os.Chdir(r.topLevel)
	if err != nil {
		return "", err
	}

	return filepath.Abs(fmt.Sprintf("target/%s/release", platform))
}

func (r rust) canCompile() (bool, error) {
	return canCompileWith(cargoConfig, r.topLevel)
}

func (r rust) compile() error {
	input := marecmd.Input{
		Command: fmt.Sprintf("cargo build --release --target %s", platform),
		Env:     map[string]string{"RUSTFLAGS": "-C target-feature=+crt-static"},
	}
	_, err := marecmd.RunFmtErr(input)
	return err
}

func (r rust) cleanup() error {
	return nil
}

func (r rust) currentVersion() (string, error) {
	f, err := os.Open(cargoConfig)
	if err != nil {
		return "", err
	}
	defer f.Close()

	regex, err := regexp.Compile(versionLinePattern)
	if err != nil {
		return "", err
	}

	s := bufio.NewScanner(f)
	for s.Scan() {
		line := s.Text()
		matches := regex.FindStringSubmatch(line)
		if len(matches) != 2 {
			continue
		}

		return matches[1], nil
	}

	return "", fmt.Errorf("unable to find version line in %s", cargoConfig)
}

func cargoCompiler(topLevel string) compiler {
	return rust{topLevel: topLevel}
}
