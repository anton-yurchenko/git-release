package release_test

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/anton-yurchenko/git-release/mocks"
	"github.com/anton-yurchenko/git-release/release"
	changelog "github.com/anton-yurchenko/go-changelog"
	"github.com/google/go-github/github"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/afero"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func int64P(n int64) *int64 {
	return &n
}

func TestGetSlug(t *testing.T) {
	a := assert.New(t)

	type expected struct {
		Result *release.Slug
		Error  string
	}

	type test struct {
		GitHubRepository string
		Expected         expected
	}

	suite := map[string]test{
		"Success": {
			GitHubRepository: "anton-yurchenko/git-release",
			Expected: expected{
				Result: &release.Slug{
					Owner: "anton-yurchenko",
					Name:  "git-release",
				},
				Error: "",
			},
		},
		"No Match": {
			GitHubRepository: "anton-yurchenkogit-release",
			Expected: expected{
				Result: nil,
				Error:  fmt.Sprintf("malformed GITHUB_REPOSITORY (expected '%v', received 'anton-yurchenkogit-release')", release.SlugRegex),
			},
		},
		"Empty GITHUB_REPOSITORY": {
			GitHubRepository: "",
			Expected: expected{
				Result: nil,
				Error:  "GITHUB_REPOSITORY is not defined",
			},
		},
	}

	var counter int
	for name, test := range suite {
		counter++
		t.Logf("Test Case %v/%v - %s", counter, len(suite), name)

		// prepare test case
		if err := os.Setenv("GITHUB_REPOSITORY", test.GitHubRepository); err != nil {
			t.Errorf("error preparing test case: error setting environmental variable GITHUB_REPOSITORY=%v: %v", test.GitHubRepository, err)
			continue
		}
		time.Sleep(30 * time.Millisecond)

		// test
		r, err := release.GetSlug()
		a.Equal(test.Expected.Result, r)
		if test.Expected.Error != "" || err != nil {
			a.EqualError(err, test.Expected.Error)
		}

		// cleanup
		if err := os.Unsetenv("GITHUB_REPOSITORY"); err != nil {
			t.Errorf("error cleanup: error unsetting environmental variable GITHUB_REPOSITORY: %v", err)
			continue
		}
		time.Sleep(30 * time.Millisecond)
	}
}

func TestGetReference(t *testing.T) {
	a := assert.New(t)

	type expected struct {
		Result *release.Reference
		Error  string
	}

	type test struct {
		GitHubRef string
		GitHubSha string
		Prefix    string
		Expected  expected
	}

	suite := map[string]test{
		"Success": {
			GitHubRef: "refs/tags/1.0.0",
			GitHubSha: "111",
			Prefix:    "",
			Expected: expected{
				Result: &release.Reference{
					CommitHash: "111",
					Version:    "1.0.0",
					Tag:        "1.0.0",
				},
				Error: "",
			},
		},
		"Empty GITHUB_REF": {
			GitHubRef: "",
			GitHubSha: "111",
			Prefix:    "",
			Expected: expected{
				Result: nil,
				Error:  "GITHUB_REF is not defined",
			},
		},
		"Empty GITHUB_SHA": {
			GitHubRef: "refs/tags/1.0.0",
			GitHubSha: "",
			Prefix:    "",
			Expected: expected{
				Result: nil,
				Error:  "GITHUB_SHA is not defined",
			},
		},
		"Tag with 'v' Prefix": {
			GitHubRef: "refs/tags/v1.0.0",
			GitHubSha: "111",
			Prefix:    "",
			Expected: expected{
				Result: &release.Reference{
					CommitHash: "111",
					Version:    "1.0.0",
					Tag:        "v1.0.0",
				},
				Error: "",
			},
		},
		"Tag with custom Prefix": {
			GitHubRef: "refs/tags/a1.0.0",
			GitHubSha: "111",
			Prefix:    "a",
			Expected: expected{
				Result: &release.Reference{
					CommitHash: "111",
					Version:    "1.0.0",
					Tag:        "a1.0.0",
				},
				Error: "",
			},
		},
		"Tag with Regex Prefix": {
			GitHubRef: "refs/tags/prerelease-1.0.0",
			GitHubSha: "111",
			Prefix:    "[a-z-]*",
			Expected: expected{
				Result: &release.Reference{
					CommitHash: "111",
					Version:    "1.0.0",
					Tag:        "prerelease-1.0.0",
				},
				Error: "",
			},
		},
		"Tag with not matching Regex Prefix": {
			GitHubRef: "refs/tags/prerelease-1.0.0",
			GitHubSha: "111",
			Prefix:    "[a-b]*",
			Expected: expected{
				Result: nil,
				Error:  fmt.Sprintf("malformed env.var GITHUB_REF: expected to match regex '^refs/tags/(?P<prefix>[a-b]*)%v$', got 'refs/tags/prerelease-1.0.0'", changelog.SemVerRegex),
			},
		},
		"Tag with custom Prefix and 'v' Prefix": {
			GitHubRef: "refs/tags/av1.0.0",
			GitHubSha: "111",
			Prefix:    "a",
			Expected: expected{
				Result: nil,
				Error:  fmt.Sprintf("malformed env.var GITHUB_REF: expected to match regex '^refs/tags/(?P<prefix>a)%v$', got 'refs/tags/av1.0.0'", changelog.SemVerRegex),
			},
		},
		"Invalid Semver": {
			GitHubRef: "refs/tags/1",
			GitHubSha: "111",
			Prefix:    "",
			Expected: expected{
				Result: nil,
				Error:  fmt.Sprintf("malformed env.var GITHUB_REF: expected to match regex '^refs/tags/[v]?%v$', got 'refs/tags/1'", changelog.SemVerRegex),
			},
		},
		"Complex Semver": {
			GitHubRef: "refs/tags/v1.0.0-alpha-a.b-c-somethinglong+build.1-aef.1-its-okay",
			GitHubSha: "111",
			Prefix:    "",
			Expected: expected{
				Result: &release.Reference{
					CommitHash: "111",
					Version:    "1.0.0-alpha-a.b-c-somethinglong+build.1-aef.1-its-okay",
					Tag:        "v1.0.0-alpha-a.b-c-somethinglong+build.1-aef.1-its-okay",
				},
				Error: "",
			},
		},
		"Complex Semver with Custom Prefix": {
			GitHubRef: "refs/tags/1.0.01.0.0-alpha-a.b-c-somethinglong+build.1-aef.1-its-okay",
			GitHubSha: "111",
			Prefix:    "1.0.0",
			Expected: expected{
				Result: &release.Reference{
					CommitHash: "111",
					Version:    "1.0.0-alpha-a.b-c-somethinglong+build.1-aef.1-its-okay",
					Tag:        "1.0.01.0.0-alpha-a.b-c-somethinglong+build.1-aef.1-its-okay",
				},
				Error: "",
			},
		},
	}

	var counter int
	for name, test := range suite {
		counter++
		t.Logf("Test Case %v/%v - %s", counter, len(suite), name)

		// prepare test case
		if err := os.Setenv("GITHUB_REF", test.GitHubRef); err != nil {
			t.Errorf("error preparing test case: error setting environmental variable GITHUB_REF=%v: %v", test.GitHubRef, err)
			continue
		}
		if err := os.Setenv("GITHUB_SHA", test.GitHubSha); err != nil {
			t.Errorf("error preparing test case: error setting environmental variable GITHUB_SHA=%v: %v", test.GitHubSha, err)
			continue
		}
		time.Sleep(30 * time.Millisecond)

		// test
		r, err := release.GetReference(test.Prefix)
		a.Equal(test.Expected.Result, r)
		if test.Expected.Error != "" || err != nil {
			a.EqualError(err, test.Expected.Error)
		}

		// cleanup
		if err := os.Unsetenv("GITHUB_REF"); err != nil {
			t.Errorf("error cleanup: error unsetting environmental variable GITHUB_REF: %v", err)
			continue
		}
		if err := os.Unsetenv("GITHUB_SHA"); err != nil {
			t.Errorf("error cleanup: error unsetting environmental variable GITHUB_SHA: %v", err)
			continue
		}
		time.Sleep(30 * time.Millisecond)
	}
}

func TestGetRelease(t *testing.T) {
	a := assert.New(t)
	roBase := afero.NewReadOnlyFs(afero.NewOsFs())
	fs := afero.NewCopyOnWriteFs(roBase, afero.NewMemMapFs())

	type expected struct {
		Result *release.Release
		Error  string
	}

	type test struct {
		GitHubRef        string
		GitHubSha        string
		GitHubRepository string
		TagPrefix        string
		DraftRelease     string
		PreRelease       string
		Name             string
		NamePrefix       string
		NameSuffix       string
		Files            []string
		Expected         expected
	}

	suite := map[string]test{
		"Success": {
			GitHubRef:        "refs/tags/1.0.0",
			GitHubSha:        "111",
			GitHubRepository: "anton-yurchenko/git-release",
			TagPrefix:        "",
			DraftRelease:     "false",
			PreRelease:       "false",
			Name:             "",
			NamePrefix:       "",
			NameSuffix:       "",
			Files:            []string{"file1", "file2"},
			Expected: expected{
				Result: &release.Release{
					Name: "1.0.0",
					Slug: &release.Slug{
						Owner: "anton-yurchenko",
						Name:  "git-release",
					},
					Reference: &release.Reference{
						CommitHash: "111",
						Tag:        "1.0.0",
						Version:    "1.0.0",
					},
					Draft:      false,
					PreRelease: false,
					Assets: &[]release.Asset{
						{
							Name: "file1",
							Path: "file1",
						},
						{
							Name: "file2",
							Path: "file2",
						},
					},
				},
				Error: "",
			},
		},
		"Tag Prefix": {
			GitHubRef:        "refs/tags/abc1.0.0",
			GitHubSha:        "111",
			GitHubRepository: "anton-yurchenko/git-release",
			TagPrefix:        "abc",
			DraftRelease:     "false",
			PreRelease:       "false",
			Name:             "",
			NamePrefix:       "",
			NameSuffix:       "",
			Files:            []string{},
			Expected: expected{
				Result: &release.Release{
					Name: "abc1.0.0",
					Slug: &release.Slug{
						Owner: "anton-yurchenko",
						Name:  "git-release",
					},
					Reference: &release.Reference{
						CommitHash: "111",
						Tag:        "abc1.0.0",
						Version:    "1.0.0",
					},
					Draft:      false,
					PreRelease: false,
					Assets:     &[]release.Asset{},
				},
				Error: "",
			},
		},
		"Draft Release": {
			GitHubRef:        "refs/tags/1.0.0",
			GitHubSha:        "111",
			GitHubRepository: "anton-yurchenko/git-release",
			TagPrefix:        "",
			DraftRelease:     "true",
			PreRelease:       "false",
			Name:             "",
			NamePrefix:       "",
			NameSuffix:       "",
			Files:            []string{},
			Expected: expected{
				Result: &release.Release{
					Name: "1.0.0",
					Slug: &release.Slug{
						Owner: "anton-yurchenko",
						Name:  "git-release",
					},
					Reference: &release.Reference{
						CommitHash: "111",
						Tag:        "1.0.0",
						Version:    "1.0.0",
					},
					Draft:      true,
					PreRelease: false,
					Assets:     &[]release.Asset{},
				},
				Error: "",
			},
		},
		"Pre Release": {
			GitHubRef:        "refs/tags/1.0.0",
			GitHubSha:        "111",
			GitHubRepository: "anton-yurchenko/git-release",
			TagPrefix:        "",
			DraftRelease:     "false",
			PreRelease:       "true",
			Name:             "",
			NamePrefix:       "",
			NameSuffix:       "",
			Files:            []string{},
			Expected: expected{
				Result: &release.Release{
					Name: "1.0.0",
					Slug: &release.Slug{
						Owner: "anton-yurchenko",
						Name:  "git-release",
					},
					Reference: &release.Reference{
						CommitHash: "111",
						Tag:        "1.0.0",
						Version:    "1.0.0",
					},
					Draft:      false,
					PreRelease: true,
					Assets:     &[]release.Asset{},
				},
				Error: "",
			},
		},
		"Invalid Semver": {
			GitHubRef:        "refs/tags/1",
			GitHubSha:        "111",
			GitHubRepository: "anton-yurchenko/git-release",
			TagPrefix:        "",
			DraftRelease:     "false",
			PreRelease:       "false",
			Name:             "",
			NamePrefix:       "",
			NameSuffix:       "",
			Files:            []string{},
			Expected: expected{
				Result: nil,
				Error:  fmt.Sprintf("error retrieving source code reference (control tag prefix via env.var TAG_PREFIX_REGEX): malformed env.var GITHUB_REF: expected to match regex '^refs/tags/[v]?%v$', got 'refs/tags/1'", changelog.SemVerRegex),
			},
		},
		"Invalid Slug": {
			GitHubRef:        "refs/tags/1.0.0",
			GitHubSha:        "111",
			GitHubRepository: "anton-yurchenkogit-release",
			TagPrefix:        "",
			DraftRelease:     "false",
			PreRelease:       "false",
			Name:             "",
			NamePrefix:       "",
			NameSuffix:       "",
			Files:            []string{},
			Expected: expected{
				Result: nil,
				Error:  fmt.Sprintf("error retrieving repository slug: malformed GITHUB_REPOSITORY (expected '%v', received 'anton-yurchenkogit-release')", release.SlugRegex),
			},
		},
		"Custom Name": {
			GitHubRef:        "refs/tags/1.0.0",
			GitHubSha:        "111",
			GitHubRepository: "anton-yurchenko/git-release",
			TagPrefix:        "",
			DraftRelease:     "false",
			PreRelease:       "false",
			Name:             "name",
			NamePrefix:       "",
			NameSuffix:       "",
			Files:            []string{},
			Expected: expected{
				Result: &release.Release{
					Name: "name",
					Slug: &release.Slug{
						Owner: "anton-yurchenko",
						Name:  "git-release",
					},
					Reference: &release.Reference{
						CommitHash: "111",
						Tag:        "1.0.0",
						Version:    "1.0.0",
					},
					Draft:      false,
					PreRelease: false,
					Assets:     &[]release.Asset{},
				},
				Error: "",
			},
		},
		"Custom Name Prefix": {
			GitHubRef:        "refs/tags/1.0.0",
			GitHubSha:        "111",
			GitHubRepository: "anton-yurchenko/git-release",
			TagPrefix:        "",
			DraftRelease:     "false",
			PreRelease:       "false",
			Name:             "",
			NamePrefix:       "prefix: ",
			NameSuffix:       "",
			Files:            []string{},
			Expected: expected{
				Result: &release.Release{
					Name: "prefix: 1.0.0",
					Slug: &release.Slug{
						Owner: "anton-yurchenko",
						Name:  "git-release",
					},
					Reference: &release.Reference{
						CommitHash: "111",
						Tag:        "1.0.0",
						Version:    "1.0.0",
					},
					Draft:      false,
					PreRelease: false,
					Assets:     &[]release.Asset{},
				},
				Error: "",
			},
		},
		"Custom Name Suffix": {
			GitHubRef:        "refs/tags/1.0.0",
			GitHubSha:        "111",
			GitHubRepository: "anton-yurchenko/git-release",
			TagPrefix:        "",
			DraftRelease:     "false",
			PreRelease:       "false",
			Name:             "",
			NamePrefix:       "",
			NameSuffix:       " suffix",
			Files:            []string{},
			Expected: expected{
				Result: &release.Release{
					Name: "1.0.0 suffix",
					Slug: &release.Slug{
						Owner: "anton-yurchenko",
						Name:  "git-release",
					},
					Reference: &release.Reference{
						CommitHash: "111",
						Tag:        "1.0.0",
						Version:    "1.0.0",
					},
					Draft:      false,
					PreRelease: false,
					Assets:     &[]release.Asset{},
				},
				Error: "",
			},
		},
		"Custom Name Prefix and Suffix": {
			GitHubRef:        "refs/tags/v1.0.0",
			GitHubSha:        "111",
			GitHubRepository: "anton-yurchenko/git-release",
			TagPrefix:        "",
			DraftRelease:     "false",
			PreRelease:       "false",
			Name:             "",
			NamePrefix:       "prefix: ",
			NameSuffix:       " suffix",
			Files:            []string{},
			Expected: expected{
				Result: &release.Release{
					Name: "prefix: v1.0.0 suffix",
					Slug: &release.Slug{
						Owner: "anton-yurchenko",
						Name:  "git-release",
					},
					Reference: &release.Reference{
						CommitHash: "111",
						Tag:        "v1.0.0",
						Version:    "1.0.0",
					},
					Draft:      false,
					PreRelease: false,
					Assets:     &[]release.Asset{},
				},
				Error: "",
			},
		},
	}

	var counter int
	for name, test := range suite {
		counter++
		t.Logf("Test Case %v/%v - %s", counter, len(suite), name)

		// prepare test case
		for _, f := range test.Files {
			if err := afero.WriteFile(fs, f, []byte(""), 0644); err != nil {
				t.Errorf("error preparing test case: error creating file %v: %v", f, err)
				continue
			}
		}

		if err := os.Setenv("GITHUB_REF", test.GitHubRef); err != nil {
			t.Errorf("error preparing test case: error setting environmental variable GITHUB_REF=%v: %v", test.GitHubRef, err)
			continue
		}
		if err := os.Setenv("GITHUB_SHA", test.GitHubSha); err != nil {
			t.Errorf("error preparing test case: error setting environmental variable GITHUB_SHA=%v: %v", test.GitHubSha, err)
			continue
		}
		if err := os.Setenv("GITHUB_REPOSITORY", test.GitHubRepository); err != nil {
			t.Errorf("error preparing test case: error setting environmental variable GITHUB_REPOSITORY=%v: %v", test.GitHubRepository, err)
			continue
		}
		if err := os.Setenv("DRAFT_RELEASE", test.DraftRelease); err != nil {
			t.Errorf("error preparing test case: error setting environmental variable DRAFT_RELEASE=%v: %v", test.DraftRelease, err)
			continue
		}
		if err := os.Setenv("PRE_RELEASE", test.PreRelease); err != nil {
			t.Errorf("error preparing test case: error setting environmental variable PRE_RELEASE=%v: %v", test.PreRelease, err)
			continue
		}
		time.Sleep(30 * time.Millisecond)

		// test
		r, err := release.GetRelease(fs, test.Files, test.TagPrefix, test.Name, test.NamePrefix, test.NameSuffix)
		a.Equal(test.Expected.Result, r)
		if test.Expected.Error != "" || err != nil {
			a.EqualError(err, test.Expected.Error)
		}

		// cleanup
		for _, f := range test.Files {
			if err := fs.Remove(f); err != nil {
				t.Errorf("error cleanup: error removing file %v: %v", f, err)
			}
		}

		if err := os.Unsetenv("GITHUB_REF"); err != nil {
			t.Errorf("error cleanup: error unsetting environmental variable GITHUB_REF: %v", err)
			continue
		}
		if err := os.Unsetenv("GITHUB_SHA"); err != nil {
			t.Errorf("error cleanup: error unsetting environmental variable GITHUB_SHA: %v", err)
			continue
		}
		if err := os.Unsetenv("GITHUB_REPOSITORY"); err != nil {
			t.Errorf("error cleanup: error unsetting environmental variable GITHUB_REPOSITORY: %v", err)
			continue
		}
		if err := os.Unsetenv("DRAFT_RELEASE"); err != nil {
			t.Errorf("error cleanup: error unsetting environmental variable DRAFT_RELEASE: %v", err)
			continue
		}
		if err := os.Unsetenv("PRE_RELEASE"); err != nil {
			t.Errorf("error cleanup: error unsetting environmental variable PRE_RELEASE: %v", err)
			continue
		}
		time.Sleep(30 * time.Millisecond)
	}
}

func TestPublish(t *testing.T) {
	a := assert.New(t)
	log.SetOutput(ioutil.Discard)
	fs := afero.NewOsFs()

	type createReleaseMock struct {
		Output *github.RepositoryRelease
		Error  error
	}

	type test struct {
		Release                *release.Release
		CreateReleaseMock      createReleaseMock
		UploadReleaseAssetMock []error
		ExpectedError          string
	}

	suite := map[string]test{
		"Without Assets": {
			Release: &release.Release{
				Name: "1.0.0",
				Slug: &release.Slug{
					Owner: "anton-yurchenko",
					Name:  "git-release",
				},
				Reference: &release.Reference{
					CommitHash: "111",
					Tag:        "1.0.0",
					Version:    "1.0.0",
				},
				Draft:      false,
				PreRelease: false,
				Assets:     nil,
				Changelog:  "changelog",
			},
			CreateReleaseMock: createReleaseMock{
				Output: nil,
				Error:  nil,
			},
			UploadReleaseAssetMock: []error{},
			ExpectedError:          "",
		},
		"With Assets": {
			Release: &release.Release{
				Name: "1.0.0",
				Slug: &release.Slug{
					Owner: "anton-yurchenko",
					Name:  "git-release",
				},
				Reference: &release.Reference{
					CommitHash: "111",
					Tag:        "1.0.0",
					Version:    "1.0.0",
				},
				Draft:      false,
				PreRelease: false,
				Assets: &[]release.Asset{
					{
						Name: "file1",
						Path: "file1",
					},
					{
						Name: "file2",
						Path: "file2",
					},
				},
				Changelog: "changelog",
			},
			CreateReleaseMock: createReleaseMock{
				Output: &github.RepositoryRelease{
					ID: int64P(2),
				},
				Error: nil,
			},
			UploadReleaseAssetMock: []error{
				nil,
				nil,
			},
			ExpectedError: "",
		},
		"Error Creating Release": {
			Release: &release.Release{
				Name: "1.0.0",
				Slug: &release.Slug{
					Owner: "anton-yurchenko",
					Name:  "git-release",
				},
				Reference: &release.Reference{
					CommitHash: "111",
					Tag:        "1.0.0",
					Version:    "1.0.0",
				},
				Draft:      false,
				PreRelease: false,
				Assets:     nil,
				Changelog:  "changelog",
			},
			CreateReleaseMock: createReleaseMock{
				Output: nil,
				Error:  errors.New("reason"),
			},
			UploadReleaseAssetMock: []error{},
			ExpectedError:          "reason",
		},
		"Error Uploading Assets": {
			Release: &release.Release{
				Name: "1.0.0",
				Slug: &release.Slug{
					Owner: "anton-yurchenko",
					Name:  "git-release",
				},
				Reference: &release.Reference{
					CommitHash: "111",
					Tag:        "1.0.0",
					Version:    "1.0.0",
				},
				Draft:      false,
				PreRelease: false,
				Assets: &[]release.Asset{
					{
						Name: "file1",
						Path: "file1",
					},
				},
				Changelog: "changelog",
			},
			CreateReleaseMock: createReleaseMock{
				Output: &github.RepositoryRelease{
					ID: int64P(2),
				},
				Error: nil,
			},
			UploadReleaseAssetMock: []error{
				errors.New("reason"),
			},
			ExpectedError: "error uploading assets",
		},
	}

	var counter int
main:
	for name, test := range suite {
		counter++
		t.Logf("Test Case %v/%v - %s", counter, len(suite), name)

		// prepare test case
		if test.Release.Assets != nil {
			for _, asset := range *test.Release.Assets {
				if err := afero.WriteFile(fs, asset.Path, []byte(""), 0644); err != nil {
					t.Errorf("error preparing test case: error creating file %v: %v", asset.Path, err)
					continue main
				}
			}
		}
		time.Sleep(30 * time.Millisecond)

		// test
		m := new(mocks.Client)

		m.On("CreateRelease",
			context.Background(),
			test.Release.Slug.Owner,
			test.Release.Slug.Name,
			&github.RepositoryRelease{
				Name:            &test.Release.Name,
				TagName:         &test.Release.Reference.Tag,
				TargetCommitish: &test.Release.Reference.CommitHash,
				Body:            &test.Release.Changelog,
				Draft:           &test.Release.Draft,
				Prerelease:      &test.Release.PreRelease,
			}).Return(test.CreateReleaseMock.Output, nil, test.CreateReleaseMock.Error).Once()

		if test.Release.Assets != nil {
			for i, asset := range *test.Release.Assets {
				m.On("UploadReleaseAsset",
					context.Background(),
					test.Release.Slug.Owner,
					test.Release.Slug.Name,
					func() int64 {
						if test.CreateReleaseMock.Output != nil {
							return *test.CreateReleaseMock.Output.ID
						} else {
							return int64(0)
						}
					}(),
					&github.UploadOptions{
						Name: strings.ReplaceAll(asset.Name, "/", "-"),
					},
					mock.AnythingOfType("*os.File")).Return(nil, nil, test.UploadReleaseAssetMock[i]).Once()
			}
		}

		err := test.Release.Publish(m)
		if test.ExpectedError != "" || err != nil {
			a.EqualError(err, test.ExpectedError)
		}

		// cleanup
		if test.Release.Assets != nil {
			for _, asset := range *test.Release.Assets {
				if err := fs.Remove(asset.Path); err != nil {
					t.Errorf("error cleanup: error removing file %v: %v", asset.Path, err)
					continue main
				}
			}
		}
		time.Sleep(30 * time.Millisecond)
	}
}
