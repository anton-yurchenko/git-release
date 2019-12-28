package app_test

import (
	"os"
	"testing"

	"github.com/anton-yurchenko/git-release/internal/app"
	"github.com/anton-yurchenko/git-release/internal/pkg/asset"
	"github.com/anton-yurchenko/git-release/internal/pkg/release"
	"github.com/anton-yurchenko/git-release/internal/pkg/repository"
	"github.com/anton-yurchenko/git-release/mocks"
	"github.com/anton-yurchenko/git-release/pkg/changelog"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
)

func TestGetConfig(t *testing.T) {
	assert := assert.New(t)
	log.SetLevel(log.FatalLevel)
	fs := afero.NewMemMapFs()
	rel := new(release.Release)
	rel.Changes = new(changelog.Changes)

	err := os.Setenv("GITHUB_WORKSPACE", ".")
	assert.Equal(nil, err, "preparation: error setting env.var 'GITHUB_WORKSPACE'")
	file, err := fs.Create("CHANGELOG.md")
	file.Close()
	assert.Equal(nil, err, "preparation: error creating test file 'CHANGELOG.md'")

	// TEST: missing GITHUB_TOKEN
	_, _, err = app.GetConfig(rel, rel.Changes, fs, []string{})

	assert.EqualError(err, "env.var 'GITHUB_TOKEN' not defined")

	// TEST: token
	err = os.Setenv("GITHUB_TOKEN", "value")
	assert.Equal(nil, err, "preparation: error setting env.var 'GITHUB_TOKEN'")

	expectedToken := "value"

	_, token, err := app.GetConfig(rel, rel.Changes, fs, []string{})

	assert.Equal(expectedToken, token)
	assert.Equal(nil, err)

	// TEST: Configuration: AllowEmptyChangelog
	err = os.Setenv("ALLOW_EMPTY_CHANGELOG", "true")
	assert.Equal(nil, err, "preparation: error setting env.var 'ALLOW_EMPTY_CHANGELOG'")

	rel = new(release.Release)
	rel.Changes = new(changelog.Changes)
	expectedConfig := &app.Configuration{
		AllowEmptyChangelog: true,
	}
	expectedToken = "value"

	config, token, err := app.GetConfig(rel, rel.Changes, fs, []string{})

	assert.Equal(expectedConfig, config)
	assert.Equal(expectedToken, token)
	assert.Equal(nil, err)

	// TEST: Draft setting
	err = os.Setenv("DRAFT_RELEASE", "true")
	assert.Equal(nil, err, "preparation: error setting env.var 'DRAFT_RELEASE'")

	rel = new(release.Release)
	rel.Changes = new(changelog.Changes)
	expectedConfig = &app.Configuration{
		AllowEmptyChangelog: true,
	}
	expectedToken = "value"
	expectedRelease := &release.Release{
		Draft: true,
		Changes: &changelog.Changes{
			File: "./CHANGELOG.md",
		},
		Assets: []asset.Asset{},
	}

	config, token, err = app.GetConfig(rel, rel.Changes, fs, []string{})

	assert.Equal(expectedConfig, config)
	assert.Equal(expectedToken, token)
	assert.Equal(nil, err)
	assert.Equal(expectedRelease, rel)

	// TEST: PreRelease setting
	err = os.Setenv("PRE_RELEASE", "true")
	assert.Equal(nil, err, "preparation: error setting env.var 'PRE_RELEASE'")

	rel = new(release.Release)
	rel.Changes = new(changelog.Changes)
	expectedConfig = &app.Configuration{
		AllowEmptyChangelog: true,
	}
	expectedToken = "value"
	expectedRelease = &release.Release{
		Draft:      true,
		PreRelease: true,
		Changes: &changelog.Changes{
			File: "./CHANGELOG.md",
		},
		Assets: []asset.Asset{},
	}

	config, token, err = app.GetConfig(rel, rel.Changes, fs, []string{})

	assert.Equal(expectedConfig, config)
	assert.Equal(expectedToken, token)
	assert.Equal(nil, err)
	assert.Equal(expectedRelease, rel)
}

func TestHydrate(t *testing.T) {
	assert := assert.New(t)

	m := new(mocks.Repository)
	v := "1.0.0"

	m.On("ReadProjectName").Return(nil).Once()
	m.On("ReadCommitHash").Return(nil).Once()
	m.On("ReadTag", &v).Return(nil).Once()

	err := app.Hydrate(m, &v)

	assert.Equal(nil, err)
}

func TestGetReleaseBody(t *testing.T) {
	assert := assert.New(t)

	// TEST: valid content
	m := new(mocks.Changelog)
	fs := afero.NewMemMapFs()
	conf := new(app.Configuration)

	m.On("ReadChanges", fs).Return(nil).Once()
	m.On("GetBody").Return("content").Once()

	err := conf.GetReleaseBody(m, fs)

	assert.Equal(nil, err)

	// TEST: empty content and AllowEmptyChangelog is enabled
	m = new(mocks.Changelog)
	fs = afero.NewMemMapFs()
	conf = &app.Configuration{
		AllowEmptyChangelog: false,
	}

	m.On("ReadChanges", fs).Return(nil).Once()
	m.On("GetBody").Return("").Once()

	err = conf.GetReleaseBody(m, fs)

	assert.EqualError(err, "changelog does not contain changes for requested project version")
}

func TestPublish(t *testing.T) {
	assert := assert.New(t)
	log.SetLevel(log.FatalLevel)

	// TEST: no exec errors
	m := new(mocks.Release)
	svc := new(mocks.GitHub)
	repo := new(repository.Repository)
	conf := app.Configuration{}

	m.On("Publish").Return(nil).Once()
	m.On("GetAssets").Return(nil)

	err := conf.Publish(repo, m, svc)

	assert.Equal(nil, err)
}
