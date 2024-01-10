package internal

import "os"

type compiler interface {
	assetDir() (string, error)
	assetFile(string, string) string
	canCompile() (bool, error)
	cleanup() error
	compile() error
	currentVersion() (string, error)
}

func canCompileWith(file, topLevel string) (bool, error) {
	err := os.Chdir(topLevel)
	if err != nil {
		return false, err
	}

	_, err = os.Stat(file)
	if os.IsNotExist(err) {
		return false, nil
	} else if err == nil {
		return true, nil
	}

	return false, err
}
