package app_test

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/anton-yurchenko/git-release/internal/app"
	"github.com/anton-yurchenko/git-release/internal/pkg/asset"
	"github.com/anton-yurchenko/git-release/internal/pkg/release"
	"github.com/anton-yurchenko/git-release/internal/pkg/repository"
	"github.com/anton-yurchenko/git-release/mocks"
	"github.com/anton-yurchenko/git-release/pkg/changelog"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
)

func TestGetConfig(t *testing.T) {
	assert := assert.New(t)
	log.SetLevel(log.FatalLevel)
	fs := afero.NewMemMapFs()

	type expected struct {
		Config  *app.Configuration
		Token   string
		Error   string
		Release *release.Release
	}

	type test struct {
		EnvVars   map[string]string
		Changelog string
		Expected  expected
	}

	suite := map[string]test{
		"Missing Required Env.Var: GITHUB_TOKEN": {
			EnvVars: map[string]string{
				"GITHUB_WORKSPACE": ".",
				"GITHUB_TOKEN":     "",
			},
			Changelog: "CHANGELOG.md",
			Expected: expected{
				Config: new(app.Configuration),
				Token:  "",
				Error:  "'GITHUB_TOKEN' is not defined",
			},
		},
		"Required Env.Var: GITHUB_TOKEN": {
			EnvVars: map[string]string{
				"GITHUB_WORKSPACE": ".",
				"GITHUB_TOKEN":     "abc123",
			},
			Changelog: "CHANGELOG.md",
			Expected: expected{
				Config: new(app.Configuration),
				Token:  "abc123",
			},
		},
		"Configuration: ALLOW_EMPTY_CHANGELOG": {
			EnvVars: map[string]string{
				"GITHUB_WORKSPACE":      ".",
				"GITHUB_TOKEN":          "token",
				"ALLOW_EMPTY_CHANGELOG": "true",
			},
			Changelog: "CHANGELOG.md",
			Expected: expected{
				Config: &app.Configuration{
					AllowEmptyChangelog: true,
				},
				Token: "token",
			},
		},
		"Configuration: ALLOW_TAG_PREFIX": {
			EnvVars: map[string]string{
				"GITHUB_WORKSPACE": ".",
				"GITHUB_TOKEN":     "token",
				"ALLOW_TAG_PREFIX": "true",
			},
			Changelog: "CHANGELOG.md",
			Expected: expected{
				Config: &app.Configuration{
					AllowTagPrefix: true,
				},
				Token: "token",
			},
		},
		"Configuration: RELEASE_NAME": {
			EnvVars: map[string]string{
				"GITHUB_WORKSPACE": ".",
				"GITHUB_TOKEN":     "token",
				"RELEASE_NAME":     "text",
			},
			Changelog: "CHANGELOG.md",
			Expected: expected{
				Config: &app.Configuration{
					ReleaseName: "text",
				},
				Token: "token",
			},
		},
		"Configuration: RELEASE_NAME_PREFIX": {
			EnvVars: map[string]string{
				"GITHUB_WORKSPACE":    ".",
				"GITHUB_TOKEN":        "token",
				"RELEASE_NAME_PREFIX": "text",
			},
			Changelog: "CHANGELOG.md",
			Expected: expected{
				Config: &app.Configuration{
					ReleaseNamePrefix: "text",
				},
				Token: "token",
			},
		},
		"Configuration: RELEASE_NAME_POSTFIX": {
			EnvVars: map[string]string{
				"GITHUB_WORKSPACE":     ".",
				"GITHUB_TOKEN":         "token",
				"RELEASE_NAME_POSTFIX": "text",
			},
			Changelog: "CHANGELOG.md",
			Expected: expected{
				Config: &app.Configuration{
					ReleaseNamePostfix: "text",
				},
				Token: "token",
			},
		},
		"Configuration: RELEASE_NAME_PREFIX & RELEASE_NAME_POSTFIX": {
			EnvVars: map[string]string{
				"GITHUB_WORKSPACE":     ".",
				"GITHUB_TOKEN":         "token",
				"RELEASE_NAME_PREFIX":  "text",
				"RELEASE_NAME_POSTFIX": "text",
			},
			Changelog: "CHANGELOG.md",
			Expected: expected{
				Config: &app.Configuration{
					ReleaseNamePrefix:  "text",
					ReleaseNamePostfix: "text",
				},
				Token: "token",
			},
		},
		"Configuration: DRAFT_RELEASE": {
			EnvVars: map[string]string{
				"GITHUB_WORKSPACE": ".",
				"GITHUB_TOKEN":     "token",
				"DRAFT_RELEASE":    "true",
			},
			Changelog: "CHANGELOG.md",
			Expected: expected{
				Config: new(app.Configuration),
				Token:  "token",
				Release: &release.Release{
					Assets: make([]asset.Asset, 0),
					Draft:  true,
					Changes: &changelog.Changes{
						File: "./CHANGELOG.md",
					},
				},
			},
		},
		"Configuration: PRE_RELEASE": {
			EnvVars: map[string]string{
				"GITHUB_WORKSPACE": ".",
				"GITHUB_TOKEN":     "token",
				"PRE_RELEASE":      "true",
			},
			Changelog: "CHANGELOG.md",
			Expected: expected{
				Config: new(app.Configuration),
				Token:  "token",
				Release: &release.Release{
					Assets:     make([]asset.Asset, 0),
					PreRelease: true,
					Changes: &changelog.Changes{
						File: "./CHANGELOG.md",
					},
				},
			},
		},
		"Configuration: Ignore Changelog": {
			EnvVars: map[string]string{
				"GITHUB_WORKSPACE": ".",
				"GITHUB_TOKEN":     "token",
				"CHANGELOG_FILE":   "none",
			},
			Changelog: "none",
			Expected: expected{
				Config: &app.Configuration{
					IgnoreChangelog: true,
				},
				Token: "token",
				Release: &release.Release{
					Assets:  make([]asset.Asset, 0),
					Changes: new(changelog.Changes),
				},
			},
		},
	}

	var counter int
	for name, test := range suite {
		counter++
		t.Logf("Test Case %v/%v - %s", counter, len(suite), name)

		// preperations
		if test.Changelog != "" {
			_, err := fs.Create(test.Changelog)
			assert.Equal(nil, err, fmt.Sprintf("preparation: error creating test file '%v'", test.Changelog))
			time.Sleep(5 * time.Millisecond)
		}

		for variable, value := range test.EnvVars {
			err := os.Setenv(variable, value)
			assert.Equal(nil, err, fmt.Sprintf("preparation: error setting environment variable '%v=%v'", variable, value))
		}

		// test
		rel := new(release.Release)
		rel.Changes = new(changelog.Changes)

		config, token, err := app.GetConfig(rel, rel.Changes, fs, []string{})

		assert.Equal(test.Expected.Config, config)
		assert.Equal(test.Expected.Token, token)
		if test.Expected.Error != "" {
			assert.EqualError(err, test.Expected.Error)
		}
		if test.Expected.Release != nil {
			assert.Equal(test.Expected.Release, rel)
		}

		// cleanup
		for variable := range test.EnvVars {
			err := os.Unsetenv(variable)
			assert.Equal(nil, err, fmt.Sprintf("preparation: error unsetting environment variable '%v'", variable))
		}

		if test.Changelog != "" {
			os.Remove(test.Changelog)
			time.Sleep(5 * time.Millisecond)
		}
	}
}

func TestHydrate(t *testing.T) {
	assert := assert.New(t)

	type test struct {
		Config               app.Configuration
		Release              release.Release
		Tag                  string
		ReadProjectNameError error
		ReadCommitHashError  error
		ReadTagError         error
		GetTagResult         string
		ExpectedError        string
	}

	suite := map[string]test{
		"Functionality": {
			Config: app.Configuration{},
			Release: release.Release{
				Changes: new(changelog.Changes),
			},
			Tag:                  "v1.0.0",
			ReadProjectNameError: nil,
			ReadCommitHashError:  nil,
			ReadTagError:         nil,
			ExpectedError:        "",
		},
		"ReadProjectName Error": {
			Config: app.Configuration{},
			Release: release.Release{
				Changes: new(changelog.Changes),
			},
			Tag:                  "v1.0.0",
			ReadProjectNameError: errors.New("error"),
			ReadCommitHashError:  nil,
			ReadTagError:         nil,
			ExpectedError:        "error",
		},
		"ReadCommitHash Error": {
			Config: app.Configuration{},
			Release: release.Release{
				Changes: new(changelog.Changes),
			},
			Tag:                  "v1.0.0",
			ReadProjectNameError: nil,
			ReadCommitHashError:  errors.New("error"),
			ReadTagError:         nil,
			ExpectedError:        "error",
		},
		"ReadTag Error": {
			Config: app.Configuration{},
			Release: release.Release{
				Changes: new(changelog.Changes),
			},
			Tag:                  "v1.0.0",
			ReadProjectNameError: nil,
			ReadCommitHashError:  nil,
			ReadTagError:         errors.New("error"),
			ExpectedError:        "error",
		},
		"Empty Release Name": {
			Config: app.Configuration{},
			Release: release.Release{
				Changes: new(changelog.Changes),
			},
			Tag:                  "v1.0.0",
			ReadProjectNameError: nil,
			ReadCommitHashError:  nil,
			ReadTagError:         nil,
			ExpectedError:        "",
		},
		"Release Name with Prefix": {
			Config: app.Configuration{
				ReleaseNamePrefix: "Prefix",
			},
			Release: release.Release{
				Changes: new(changelog.Changes),
			},
			Tag:                  "v1.0.0",
			ReadProjectNameError: nil,
			ReadCommitHashError:  nil,
			ReadTagError:         nil,
			ExpectedError:        "",
		},
		"Release Name with Postfix": {
			Config: app.Configuration{
				ReleaseNamePostfix: "Postfix",
			},
			Release: release.Release{
				Changes: new(changelog.Changes),
			},
			Tag:                  "v1.0.0",
			ReadProjectNameError: nil,
			ReadCommitHashError:  nil,
			ReadTagError:         nil,
			ExpectedError:        "",
		},
		"Release Name with Prefix and Postfix": {
			Config: app.Configuration{
				ReleaseNamePrefix:  "Prefix",
				ReleaseNamePostfix: "Postfix",
			},
			Release: release.Release{
				Changes: new(changelog.Changes),
			},
			Tag:                  "v1.0.0",
			ReadProjectNameError: nil,
			ReadCommitHashError:  nil,
			ReadTagError:         nil,
			ExpectedError:        "",
		},
	}

	var counter int
	for name, test := range suite {
		counter++
		t.Logf("Test Case %v/%v - %s", counter, len(suite), name)

		m := new(mocks.Repository)
		m.On("ReadProjectName").Return(test.ReadProjectNameError).Once()
		m.On("ReadCommitHash").Return(test.ReadCommitHashError).Once()
		m.On("ReadTag", &test.Release.Changes.Version, false).Return(test.ReadTagError).Once()
		m.On("GetTag").Return(&test.Tag).Once()

		err := test.Config.Hydrate(m, &test.Release.Changes.Version, &test.Release.Name)

		if test.ExpectedError != "" {
			assert.EqualError(err, test.ExpectedError)
			assert.Equal("", test.Release.Changes.Version)
			assert.Equal("", test.Release.Name)
		} else {
			assert.Equal(nil, err)

			if test.Config.ReleaseName != "" {
				assert.Equal(test.Release.Name, test.Config.ReleaseName)
			} else if test.Config.ReleaseNamePrefix != "" || test.Config.ReleaseNamePostfix != "" {
				assert.Equal(fmt.Sprintf("%v%v%v", test.Config.ReleaseNamePrefix, test.Tag, test.Config.ReleaseNamePostfix), test.Release.Name)
			} else {
				assert.Equal(test.Release.Name, test.Tag)
			}
		}
	}
}

func TestGetReleaseBody(t *testing.T) {
	assert := assert.New(t)
	fs := afero.NewMemMapFs()
	log.SetLevel(log.FatalLevel)

	type test struct {
		Config           app.Configuration
		ReadChangesError error
		GetBodyResult    string
		ExpectedError    string
	}

	suite := map[string]test{
		"Functionality": {
			Config:           app.Configuration{},
			ReadChangesError: nil,
			GetBodyResult:    "content",
			ExpectedError:    "",
		},
		"Empty Content with AllowEmptyChangelog Enabled": {
			Config: app.Configuration{
				AllowEmptyChangelog: false,
			},
			ReadChangesError: nil,
			GetBodyResult:    "",
			ExpectedError:    "changelog does not contain changes for requested project version",
		},
		"Changelog Error": {
			Config: app.Configuration{
				AllowEmptyChangelog: false,
			},
			ReadChangesError: errors.New("error"),
			GetBodyResult:    "content",
			ExpectedError:    "error",
		},
	}

	var counter int
	for name, test := range suite {
		counter++
		t.Logf("Test Case %v/%v - %s", counter, len(suite), name)

		m := new(mocks.Changelog)
		m.On("ReadChanges", fs).Return(test.ReadChangesError).Once()
		m.On("GetBody").Return(test.GetBodyResult).Once()

		err := test.Config.GetReleaseBody(m, fs)

		if test.ExpectedError != "" {
			assert.EqualError(err, test.ExpectedError)
		} else {
			assert.Equal(nil, err)
		}
	}
}

func TestPublish(t *testing.T) {
	assert := assert.New(t)
	log.SetLevel(log.FatalLevel)

	// TEST: no exec errors
	t.Log("Test Case 1/1 - Functionality")

	m := new(mocks.Release)
	svc := new(mocks.GitHub)
	repo := new(repository.Repository)
	conf := app.Configuration{}

	m.On("Publish").Return(nil).Once()
	m.On("GetAssets").Return(nil)

	err := conf.Publish(repo, m, svc)

	assert.Equal(nil, err)
}
