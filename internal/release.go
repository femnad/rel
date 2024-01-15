package internal

import (
	"context"
	"fmt"
	"os"
	"path"
	"path/filepath"

	"github.com/femnad/rel/git"
	"github.com/femnad/rel/github"
	"github.com/femnad/rel/log"
)

var (
	versionLinePattern = `version = "([0-9]+\.[0-9]+\.[0-9]+)"`
	compilerFns        = []func(string, string) compiler{
		goCompiler,
		rustCompiler,
	}
)

type config struct {
	TokenFromGH  bool   `yaml:"token_from_gh"`
	TokenCommand string `yaml:"token_command"`
	Token        string `yaml:"token"`
}

type Releaser struct {
	gh        github.GitHub
	comp      compiler
	gitClient git.Client
	owner     string
	repo      string
	topLevel  string
}

func findTopLevel() (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}

	for {
		repoPath := filepath.Join(cwd, ".git")
		_, err = os.Stat(repoPath)
		if err == nil {
			return cwd, nil
		}

		cwd = filepath.Dir(cwd)
		if cwd == filepath.Dir(cwd) {
			break
		}
	}

	return "", fmt.Errorf("unable to find top level")
}

func NewReleaser(configFile, path string) (r Releaser, err error) {
	gitClient, err := git.New(path)
	if err != nil {
		return
	}

	owner, repo := gitClient.Owner(), gitClient.Repo()
	topLevel, err := findTopLevel()
	if err != nil {
		return
	}

	var canCompile bool
	var comp compiler
	for _, fn := range compilerFns {
		comp = fn(repo, topLevel)
		canCompile, err = comp.canCompile()
		if err != nil {
			return r, fmt.Errorf("error determining compiler capability: %v", err)
		}

		if canCompile {
			log.Logger.Debugf("Using %s compiler", comp.name())

			r.comp = comp
			break
		}
	}

	if !canCompile {
		return r, fmt.Errorf("unable to find suitable compiler")
	}

	cfg, err := parseConfig(configFile)
	if err != nil {
		return
	}

	token, err := getToken(cfg)
	if err != nil {
		return
	}

	client := github.New(owner, repo, token)

	return Releaser{
		comp:      comp,
		gh:        client,
		gitClient: gitClient,
		owner:     owner,
		repo:      repo,
		topLevel:  topLevel,
	}, nil
}

func (r Releaser) ensureRelease(ctx context.Context, hash, version string) error {
	spec := github.ReleaseSpec{
		Hash:    hash,
		TagName: version,
	}
	id, err := r.gh.EnsureRelease(ctx, spec)
	if err != nil {
		return err
	}
	spec.ID = id

	log.Logger.Debug("Compiling release binary")

	err = r.comp.compile()
	if err != nil {
		return err
	}

	assetDir, err := r.comp.assetDir()
	if err != nil {
		return err
	}

	filePath := path.Join(assetDir, r.repo)
	asset := github.AssetSpec{
		ReleaseSpec: spec,
		Name:        r.comp.assetFile(version),
		Path:        filePath,
	}
	err = r.gh.UploadReleaseAsset(ctx, asset)
	if err != nil {
		return err
	}

	return r.gh.FinalizeRelease(ctx, spec)
}

func (r Releaser) Release(ctx context.Context) error {
	version, err := r.comp.currentVersion()
	if err != nil {
		return err
	}

	defer func() {
		cleanupErr := r.comp.cleanup()
		if cleanupErr != nil {
			log.Logger.Errorf("error cleaning up: %v", cleanupErr)
		}
	}()

	err = r.comp.compile()
	if err != nil {
		return err
	}

	hash, err := r.gitClient.Tag(version)
	if err != nil {
		return err
	}

	err = r.gitClient.Push()
	if err != nil {
		return err
	}

	return r.ensureRelease(ctx, hash, version)
}
