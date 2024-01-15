package github

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"

	gh "github.com/google/go-github/v57/github"

	"github.com/femnad/rel/log"
)

type GitHub struct {
	client *gh.Client
	owner  string
	repo   string
}

type AssetSpec struct {
	ReleaseSpec
	Name string
	Path string
}

type ReleaseSpec struct {
	Hash    string
	ID      int64
	TagName string
}

func New(owner, repo, token string) GitHub {
	return GitHub{
		client: gh.NewClient(nil).WithAuthToken(token),
		owner:  owner,
		repo:   repo,
	}
}

func newTrue() *bool {
	b := new(bool)
	*b = true
	return b
}

func (g GitHub) getReleaseByTag(ctx context.Context, tagName string) (*gh.RepositoryRelease, error) {
	rel, resp, err := g.client.Repositories.GetReleaseByTag(ctx, g.owner, g.repo, tagName)
	if resp != nil && resp.StatusCode == http.StatusNotFound {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	return rel, err
}

func (g GitHub) getReleaseByID(ctx context.Context, id int64) (*gh.RepositoryRelease, error) {
	rel, resp, err := g.client.Repositories.GetRelease(ctx, g.owner, g.repo, id)
	if resp != nil && resp.StatusCode == http.StatusNotFound {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	return rel, err
}

func (g GitHub) createRelease(ctx context.Context, spec ReleaseSpec) (*gh.RepositoryRelease, error) {
	makeLatest := "true"
	rel := &gh.RepositoryRelease{
		Draft:                newTrue(),
		GenerateReleaseNotes: newTrue(),
		MakeLatest:           &makeLatest,
		Name:                 &spec.TagName,
		TagName:              &spec.TagName,
		TargetCommitish:      &spec.Hash,
	}

	log.Logger.Infof("Creating release %s as draft", *rel.Name)

	rel, resp, err := g.client.Repositories.CreateRelease(ctx, g.owner, g.repo, rel)
	if err != nil {
		body, readErr := io.ReadAll(resp.Body)
		if readErr != nil {
			return rel, readErr
		}
		resp.Body.Close()
		return nil, fmt.Errorf("error creating release, body %s, error %v", string(body), err)
	}

	return rel, nil
}

func (g GitHub) EnsureRelease(ctx context.Context, spec ReleaseSpec) (int64, error) {
	tag := spec.TagName
	rel, err := g.getReleaseByTag(ctx, tag)
	if err != nil {
		return 0, err
	}

	if rel == nil {
		rel, err = g.createRelease(ctx, spec)
		if err != nil {
			return 0, err
		}
	} else if !rel.GetDraft() {
		return 0, fmt.Errorf("release %s exists but is not a draft release", *rel.Name)
	} else {
		log.Logger.Noticef("Release %s already exists as a draft release", *rel.Name)
	}

	return *rel.ID, nil
}

func (g GitHub) UploadReleaseAsset(ctx context.Context, spec AssetSpec) error {
	opts := &gh.UploadOptions{
		Name: spec.Name,
	}

	file, err := os.Open(spec.Path)
	if err != nil {
		return err
	}
	defer file.Close()

	log.Logger.Infof("Uploading release asset %s", opts.Name)

	_, resp, err := g.client.Repositories.UploadReleaseAsset(ctx, g.owner, g.repo, spec.ReleaseSpec.ID, opts, file)
	if err == nil {
		return nil
	}

	body, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		return readErr
	}
	resp.Body.Close()

	return fmt.Errorf("error uploading release asset, body: %s, error: %v", body, err)
}

func (g GitHub) FinalizeRelease(ctx context.Context, spec ReleaseSpec) error {
	rel, err := g.getReleaseByID(ctx, spec.ID)
	*rel.Draft = false

	log.Logger.Infof("Removing draft status from %s", *rel.Name)

	_, _, err = g.client.Repositories.EditRelease(ctx, g.owner, g.repo, *rel.ID, rel)
	return err
}
