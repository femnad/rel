package main

import (
	"context"
	"fmt"

	"github.com/alexflint/go-arg"

	"github.com/femnad/rel/internal"
	"github.com/femnad/rel/log"
)

const (
	name    = "rel"
	version = "0.2.1"
)

type args struct {
	ConfigFile string `arg:"-f,--file" default:"~/.config/rel/rel.yml" help:"Config file path"`
	Path       string `arg:"positional" default:"." help:"Repo path"`
}

func (args) Version() string {
	return fmt.Sprintf("%s v%s", name, version)
}

func release(ctx context.Context, configFile, path string) error {
	r, err := internal.NewReleaser(configFile, path)
	if err != nil {
		return err
	}

	return r.Release(ctx)
}

func main() {
	var parsed args
	arg.MustParse(&parsed)

	log.Init()

	err := release(context.Background(), parsed.ConfigFile, parsed.Path)
	if err != nil {
		log.Logger.Fatalf("Error creating release: %v", err)
	}
}
