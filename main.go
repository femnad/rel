package main

import (
	"context"
	"log"

	"github.com/alexflint/go-arg"

	"github.com/femnad/rel/internal"
)

type args struct {
	Path string `arg:"positional" default:"." help:"Repo path"`
}

func (args) Version() string {
	return "rel 0.1.0"
}

func release(ctx context.Context, path string) error {
	r, err := internal.NewReleaser(path)
	if err != nil {
		return err
	}

	return r.Release(ctx)
}

func main() {
	var parsed args
	arg.MustParse(&parsed)

	err := release(context.Background(), parsed.Path)
	if err != nil {
		log.Fatal(err)
	}
}
