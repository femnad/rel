package internal

import (
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/femnad/mare"
	"github.com/femnad/mare/cmd"
)

func tokenFromCmd(command string) (string, error) {
	out, err := cmd.RunFmtErr(cmd.Input{Command: command})
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(out.Stdout), nil
}

func getToken(cfg config) (string, error) {
	if cfg.TokenFromGH {
		return tokenFromCmd("gh auth token")
	} else if cfg.TokenCommand != "" {
		return tokenFromCmd(cfg.TokenCommand)
	} else if cfg.Token != "" {
		return cfg.Token, nil
	}

	return "", fmt.Errorf("unable to determine token getter command")
}

func parseConfig(configFile string) (cfg config, err error) {
	configFile = mare.ExpandUser(configFile)
	_, err = os.Stat(configFile)
	if os.IsNotExist(err) {
		return config{TokenFromGH: true}, nil
	} else if err != nil {
		return
	}

	file, err := os.Open(configFile)
	if err != nil {
		return cfg, fmt.Errorf("error opening config file %s: %v", configFile, err)
	}

	decoder := yaml.NewDecoder(file)
	err = decoder.Decode(&cfg)
	return cfg, fmt.Errorf("error deserializing config file %s: %v", configFile, err)
}
