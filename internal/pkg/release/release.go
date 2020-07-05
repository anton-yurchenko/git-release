package release

import (
	"context"
	"sync"

	"github.com/anton-yurchenko/git-release/internal/pkg/asset"
	"github.com/anton-yurchenko/git-release/internal/pkg/interfaces"
	"github.com/anton-yurchenko/git-release/internal/pkg/repository"
	"github.com/anton-yurchenko/git-release/pkg/changelog"
	"github.com/google/go-github/v32/github"
	"github.com/pkg/errors"
)

// Release represents a github.com release
type Release struct {
	Name       string
	Draft      bool
	PreRelease bool
	Assets     []asset.Asset
	Changes    *changelog.Changes
}

// Interface of 'Release'
type Interface interface {
	Publish(repository.Interface, interfaces.GitHub, chan string, chan error) error
	EnableDraft()
	EnablePreRelease()
	SetAssets([]asset.Asset)
	GetAssets() []asset.Asset
}

// Publish a new github.com release
func (r *Release) Publish(repo repository.Interface, service interfaces.GitHub, messages chan string, errs chan error) error {
	if r.Changes == nil {
		return errors.New("receiver contains a nil pointer")
	}

	// create release
	release, _, err := service.CreateRelease(
		context.Background(),
		repo.GetOwner(),
		repo.GetProject(),
		&github.RepositoryRelease{
			Name:            &r.Name,
			TagName:         repo.GetTag(),
			TargetCommitish: repo.GetCommitHash(),
			Body:            &r.Changes.Body,
			Draft:           &r.Draft,
			Prerelease:      &r.PreRelease,
		},
	)
	if err != nil {
		return err
	}

	// upload assets
	if len(r.Assets) > 0 {
		wg := new(sync.WaitGroup)

		wg.Add(len(r.Assets))

		for _, asset := range r.Assets {
			x := asset
			go x.Upload(release.GetID(), repo, service, wg, messages, errs)
		}
	}

	return nil
}

// EnableDraft sets release 'Draft' property to 'true'
func (r *Release) EnableDraft() {
	r.Draft = true
}

// EnablePreRelease sets release 'PreRelease' property to 'true'
func (r *Release) EnablePreRelease() {
	r.PreRelease = true
}

// SetAssets sets provided assets as a release assets
func (r *Release) SetAssets(a []asset.Asset) {
	r.Assets = a
}

// GetAssets returns release assets
func (r *Release) GetAssets() []asset.Asset {
	return r.Assets
}
