package internal

type compiler interface {
	assetDir() (string, error)
	assetFile(string, string) string
	canCompile() (bool, error)
	cleanup() error
	compile() error
	currentVersion() (string, error)
}
