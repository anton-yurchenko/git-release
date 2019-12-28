package release_test

import (
	"testing"

	"github.com/anton-yurchenko/git-release/internal/pkg/asset"
	"github.com/anton-yurchenko/git-release/internal/pkg/release"
	"github.com/anton-yurchenko/git-release/internal/pkg/repository"
	"github.com/anton-yurchenko/git-release/mocks"
	"github.com/anton-yurchenko/git-release/pkg/changelog"
	"github.com/google/go-github/github"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestPublish(t *testing.T) {
	assert := assert.New(t)
	log.SetLevel(log.FatalLevel)

	// TEST: successfull release without assets
	messages := make(chan string)
	errs := make(chan error)
	m := new(mocks.GitHub)
	repo := new(repository.Repository)
	rel := new(release.Release)
	rel.Changes = new(changelog.Changes)
	var id int64 = 12

	m.On("CreateRelease").Return(&github.RepositoryRelease{ID: &id}, new(github.Response), nil).Once()

	err := rel.Publish(repo, m, messages, errs)

	assert.Equal(nil, err)

	// TEST: successfull release with single asset
	m = new(mocks.GitHub)
	repo = new(repository.Repository)
	messages = make(chan string, 1)
	errs = make(chan error, 1)
	rel = &release.Release{
		Assets: []asset.Asset{
			asset.Asset{
				Name: "file1",
				Path: "release_test.go",
			},
		},
	}
	rel.Changes = new(changelog.Changes)
	id = 124

	m.On("CreateRelease").Return(&github.RepositoryRelease{ID: &id}, new(github.Response), nil).Once()
	m.On("UploadReleaseAsset").Return(new(github.ReleaseAsset), new(github.Response), nil).Once()

	err = rel.Publish(repo, m, messages, errs)

	assert.Equal(nil, err)

	// TEST: successfull release with multiple assets
	m = new(mocks.GitHub)
	repo = new(repository.Repository)
	messages = make(chan string, 2)
	errs = make(chan error, 2)
	rel = &release.Release{
		Assets: []asset.Asset{
			asset.Asset{
				Name: "file1",
				Path: "release_test.go",
			},
			asset.Asset{
				Name: "file1",
				Path: "asset_test.go",
			},
		},
	}
	rel.Changes = new(changelog.Changes)
	id = 124

	m.On("CreateRelease").Return(&github.RepositoryRelease{ID: &id}, new(github.Response), nil).Once()
	m.On("UploadReleaseAsset").Return(new(github.ReleaseAsset), new(github.Response), nil).Twice()

	err = rel.Publish(repo, m, messages, errs)

	assert.Equal(nil, err)

	// TEST: failed release with single asset, test function abort mechanism
	m = new(mocks.GitHub)
	repo = new(repository.Repository)
	messages = make(chan string, 1)
	errs = make(chan error, 1)
	rel = &release.Release{
		Assets: []asset.Asset{
			asset.Asset{
				Name: "file1",
				Path: "file-not-found.zip",
			},
		},
	}
	rel.Changes = new(changelog.Changes)
	id = 124

	m.On("CreateRelease").Return(&github.RepositoryRelease{ID: &id}, new(github.Response), nil).Once()
	m.On("UploadReleaseAsset").Return(new(github.ReleaseAsset), new(github.Response), nil).Once()

	err = rel.Publish(repo, m, messages, errs)

	assert.Equal(nil, err)

	err = <-errs

	assert.EqualError(err, "open file-not-found.zip: no such file or directory")

	// TEST: failed release creation
	m = new(mocks.GitHub)
	messages = make(chan string)
	errs = make(chan error)
	repo = new(repository.Repository)

	rel = new(release.Release)
	rel.Changes = new(changelog.Changes)
	id = 964

	m.On("CreateRelease").Return(&github.RepositoryRelease{ID: &id}, new(github.Response), errors.New("release creation failed")).Once()

	err = rel.Publish(repo, m, messages, errs)

	assert.EqualError(err, "release creation failed")
}

func TestEnableDraft(t *testing.T) {
	assert := assert.New(t)

	expected := true

	m := release.Release{
		Draft: false,
	}

	m.EnableDraft()

	assert.Equal(expected, m.Draft)
}

func TestEnablePreRelease(t *testing.T) {
	assert := assert.New(t)

	expected := true

	m := release.Release{
		PreRelease: false,
	}

	m.EnablePreRelease()

	assert.Equal(expected, m.PreRelease)
}

func TestSetAssets(t *testing.T) {
	assert := assert.New(t)

	m := new(release.Release)

	expected := []asset.Asset{
		asset.Asset{
			Name: "file1",
			Path: "release_test.go",
		},
		asset.Asset{
			Name: "file1",
			Path: "asset_test.go",
		},
	}

	m.SetAssets(expected)

	assert.Equal(expected, m.Assets)
}

func TestGetAssets(t *testing.T) {
	assert := assert.New(t)

	expected := []asset.Asset{
		asset.Asset{
			Name: "file1",
			Path: "release_test.go",
		},
		asset.Asset{
			Name: "file1",
			Path: "asset_test.go",
		},
	}

	m := release.Release{
		Assets: expected,
	}

	assert.Equal(expected, m.GetAssets())
}
