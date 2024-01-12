package git

import (
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"

	"github.com/femnad/rel/log"
)

const (
	defaultRemote = "origin"
)

var (
	gitSuffix   = ".git"
	repoPattern = "git@github.com:([^/]+)/([^/]+)"
)

type Client struct {
	gitRepo *git.Repository
	owner   string
	repo    string
}

func New(path string) (c Client, err error) {
	repo, err := git.PlainOpen(path)
	if err != nil {
		return
	}

	remote, err := repo.Remote(defaultRemote)
	if err != nil {
		return
	}

	regex, err := regexp.Compile(repoPattern)
	if err != nil {
		return
	}

	firstURL := remote.Config().URLs[0]
	matches := regex.FindStringSubmatch(firstURL)
	if len(matches) < 3 {
		return c, fmt.Errorf("unable to determine github repo")
	}

	repoName := matches[2]
	if strings.HasSuffix(repoName, gitSuffix) {
		repoName = strings.TrimSuffix(repoName, gitSuffix)
	}

	return Client{
		gitRepo: repo,
		owner:   matches[1],
		repo:    repoName,
	}, nil
}

func (c Client) Owner() string {
	return c.owner
}

func (c Client) Push() error {
	log.Logger.Debug("Pushing commits and tags")

	err := c.gitRepo.Push(&git.PushOptions{FollowTags: true})
	if err == nil || errors.Is(err, git.NoErrAlreadyUpToDate) {
		return nil
	}

	return err
}

func (c Client) Repo() string {
	return c.repo
}

func (c Client) Tag(version string) (string, error) {
	tags, err := c.gitRepo.Tags()
	if err != nil {
		return "", err
	}

	var hash plumbing.Hash
	err = tags.ForEach(func(reference *plumbing.Reference) error {
		if !hash.IsZero() {
			return nil
		}

		if version == reference.Name().Short() {
			log.Logger.Noticef("Tag %s already exists", version)
			hash = reference.Hash()
		}
		return nil
	})
	if err != nil {
		return "", err
	}

	if !hash.IsZero() {
		return hash.String(), nil
	}

	head, err := c.gitRepo.Head()
	if err != nil {
		return "", err
	}
	hash = head.Hash()

	log.Logger.Infof("Tagging %s with %s", hash, version)
	_, err = c.gitRepo.CreateTag(version, hash, nil)
	if err != nil {
		return "", err
	}

	return hash.String(), nil
}
