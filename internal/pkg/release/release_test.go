package release_test

import (
	"testing"

	"github.com/anton-yurchenko/git-release/internal/pkg/asset"
	"github.com/anton-yurchenko/git-release/internal/pkg/release"
	"github.com/anton-yurchenko/git-release/internal/pkg/repository"
	"github.com/anton-yurchenko/git-release/mocks"
	"github.com/anton-yurchenko/git-release/pkg/changelog"
	"github.com/google/go-github/v32/github"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestPublish(t *testing.T) {
	assert := assert.New(t)
	log.SetLevel(log.FatalLevel)

	type test struct {
		Release            *release.Release
		ExpectedError      string
		CreateReleaseError error
	}

	suite := map[string]test{
		"Functionality": {
			Release: &release.Release{
				Changes: &changelog.Changes{},
			},
			ExpectedError:      "",
			CreateReleaseError: nil,
		},
		"Nil Pointer Trap": {
			Release:            &release.Release{},
			ExpectedError:      "receiver contains a nil pointer",
			CreateReleaseError: nil,
		},
		"Single Asset": {
			Release: &release.Release{
				Assets: []asset.Asset{
					{
						Name: "file",
						Path: "release_test.go",
					},
				},
				Changes: &changelog.Changes{},
			},
			ExpectedError:      "",
			CreateReleaseError: nil,
		},
		"Multiple Assets": {
			Release: &release.Release{
				Assets: []asset.Asset{
					{
						Name: "file1",
						Path: "release_test.go",
					},
					{
						Name: "file2",
						Path: "asset_test.go",
					},
				},
				Changes: &changelog.Changes{},
			},
			ExpectedError:      "",
			CreateReleaseError: nil,
		},
		"Error Creating Release": {
			Release: &release.Release{
				Changes: &changelog.Changes{},
			},
			ExpectedError:      "error",
			CreateReleaseError: errors.New("error"),
		},
	}

	var counter int
	for name, test := range suite {
		counter++
		t.Logf("Test Case %v/%v - %s", counter, len(suite), name)

		var id int64 = 1
		m := new(mocks.GitHub)
		m.On("CreateRelease").Return(&github.RepositoryRelease{ID: &id}, new(github.Response), test.CreateReleaseError).Once()
		m.On("UploadReleaseAsset").Return(new(github.ReleaseAsset), new(github.Response), nil)

		err := test.Release.Publish(new(repository.Repository), m, make(chan string, len(test.Release.Assets)), make(chan error, len(test.Release.Assets)))

		if test.ExpectedError != "" {
			assert.EqualError(err, test.ExpectedError)
		} else {
			assert.Equal(nil, err)
		}
	}
}

func TestEnableDraft(t *testing.T) {
	assert := assert.New(t)
	t.Log("Test Case 1/1 - Functionality")

	expected := true

	m := release.Release{
		Draft: false,
	}

	m.EnableDraft()

	assert.Equal(expected, m.Draft)
}

func TestEnablePreRelease(t *testing.T) {
	assert := assert.New(t)
	t.Log("Test Case 1/1 - Functionality")

	expected := true

	m := release.Release{
		PreRelease: false,
	}

	m.EnablePreRelease()

	assert.Equal(expected, m.PreRelease)
}

func TestSetAssets(t *testing.T) {
	assert := assert.New(t)
	t.Log("Test Case 1/1 - Functionality")

	m := new(release.Release)

	expected := []asset.Asset{
		{
			Name: "file1",
			Path: "release_test.go",
		},
		{
			Name: "file1",
			Path: "asset_test.go",
		},
	}

	m.SetAssets(expected)

	assert.Equal(expected, m.Assets)
}

func TestGetAssets(t *testing.T) {
	assert := assert.New(t)
	t.Log("Test Case 1/1 - Functionality")

	expected := []asset.Asset{
		{
			Name: "file1",
			Path: "release_test.go",
		},
		{
			Name: "file1",
			Path: "asset_test.go",
		},
	}

	m := release.Release{
		Assets: expected,
	}

	assert.Equal(expected, m.GetAssets())
}
