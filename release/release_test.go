package release_test

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"testing"
	"time"

	"git-release/mocks"

	"git-release/release"

	changelog "github.com/anton-yurchenko/go-changelog"
	"github.com/google/go-github/v78/github"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/afero"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func stringP(n string) *string {
	return &n
}

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
		GitHubRef     string
		GitHubSha     string
		UnreleasedTag string
		Prefix        string
		Unreleased    bool
		Expected      expected
	}

	suite := map[string]test{
		// v6 extracted the version with a positional ${2}, so a capture group in
		// the user's own prefix shifted the numbering and yielded the prefix as
		// the version.
		"Tag Prefix Containing A Capture Group": {
			GitHubRef:  "refs/tags/rel-1.2.3",
			GitHubSha:  "111",
			Prefix:     "(rel|pre)-",
			Unreleased: false,
			Expected: expected{
				Result: &release.Reference{
					CommitHash: "111",
					Tag:        "rel-1.2.3",
					Version:    "1.2.3",
					Major:      "1",
					Minor:      "2",
					Patch:      "3",
				},
				Error: "",
			},
		},
		// The tag is rebuilt from every path segment after refs/tags/, so a
		// prefix containing a slash must not truncate it.
		"Tag Prefix Containing A Slash": {
			GitHubRef:  "refs/tags/pkg/v1.2.3",
			GitHubSha:  "111",
			Prefix:     "pkg/v",
			Unreleased: false,
			Expected: expected{
				Result: &release.Reference{
					CommitHash: "111",
					Tag:        "pkg/v1.2.3",
					Version:    "1.2.3",
					Major:      "1",
					Minor:      "2",
					Patch:      "3",
				},
				Error: "",
			},
		},
		// v6 compared against a hardcoded refs/tags/latest, so a workflow using a
		// custom rolling tag could trigger itself forever.
		"Triggering Loop On A Custom Unreleased Tag": {
			GitHubRef:     "refs/tags/nightly",
			GitHubSha:     "111",
			UnreleasedTag: "nightly",
			Unreleased:    true,
			Expected: expected{
				Result: nil,
				Error:  "workflow configuration error detected: trigger loop (triggering tag will be recreated and trigger the workflow again)",
			},
		},
		// ...while an ordinary release of a tag named 'latest' is legitimate:
		// v6 refused it, because the guard was not scoped to unreleased mode.
		"A Tag Named Latest Is Releasable": {
			GitHubRef:  "refs/tags/latest",
			GitHubSha:  "111",
			Unreleased: false,
			Expected: expected{
				Result: nil,
				Error:  "malformed env.var GITHUB_REF: expected to match regex '^refs/tags/(?:v?)(?P<version>" + changelog.SemVerRegex + ")$', got 'refs/tags/latest'",
			},
		},
		"Success": {
			GitHubRef:     "refs/tags/1.0.0",
			GitHubSha:     "111",
			UnreleasedTag: "",
			Prefix:        "",
			Unreleased:    false,
			Expected: expected{
				Result: &release.Reference{
					CommitHash: "111",
					Version:    "1.0.0",
					Tag:        "1.0.0",
					Major:      "1",
					Minor:      "0",
					Patch:      "0",
				},
				Error: "",
			},
		},
		"Empty GITHUB_REF": {
			GitHubRef:     "",
			GitHubSha:     "111",
			UnreleasedTag: "",
			Prefix:        "",
			Unreleased:    false,
			Expected: expected{
				Result: nil,
				Error:  "GITHUB_REF is not defined",
			},
		},
		"Empty GITHUB_SHA": {
			GitHubRef:     "refs/tags/1.0.0",
			GitHubSha:     "",
			UnreleasedTag: "",
			Prefix:        "",
			Unreleased:    false,
			Expected: expected{
				Result: nil,
				Error:  "GITHUB_SHA is not defined",
			},
		},
		"Tag with 'v' Prefix": {
			GitHubRef:     "refs/tags/v1.0.0",
			GitHubSha:     "111",
			UnreleasedTag: "",
			Prefix:        "",
			Unreleased:    false,
			Expected: expected{
				Result: &release.Reference{
					CommitHash: "111",
					Version:    "1.0.0",
					Tag:        "v1.0.0",
					Major:      "1",
					Minor:      "0",
					Patch:      "0",
				},
				Error: "",
			},
		},
		"Tag with custom Prefix": {
			GitHubRef:     "refs/tags/a1.0.0",
			GitHubSha:     "111",
			UnreleasedTag: "",
			Prefix:        "a",
			Unreleased:    false,
			Expected: expected{
				Result: &release.Reference{
					CommitHash: "111",
					Version:    "1.0.0",
					Tag:        "a1.0.0",
					Major:      "1",
					Minor:      "0",
					Patch:      "0",
				},
				Error: "",
			},
		},
		"Tag with Regex Prefix": {
			GitHubRef:     "refs/tags/prerelease-1.0.0",
			GitHubSha:     "111",
			UnreleasedTag: "",
			Prefix:        "[a-z-]*",
			Unreleased:    false,
			Expected: expected{
				Result: &release.Reference{
					CommitHash: "111",
					Version:    "1.0.0",
					Tag:        "prerelease-1.0.0",
					Major:      "1",
					Minor:      "0",
					Patch:      "0",
				},
				Error: "",
			},
		},
		"Tag with not matching Regex Prefix": {
			GitHubRef:     "refs/tags/prerelease-1.0.0",
			GitHubSha:     "111",
			UnreleasedTag: "",
			Prefix:        "[a-b]*",
			Unreleased:    false,
			Expected: expected{
				Result: nil,
				Error:  fmt.Sprintf("malformed env.var GITHUB_REF: expected to match regex '^refs/tags/(?:[a-b]*)(?P<version>%v)$', got 'refs/tags/prerelease-1.0.0'", changelog.SemVerRegex),
			},
		},
		"Tag with custom Prefix and 'v' Prefix": {
			GitHubRef:     "refs/tags/av1.0.0",
			GitHubSha:     "111",
			UnreleasedTag: "",
			Prefix:        "a",
			Unreleased:    false,
			Expected: expected{
				Result: nil,
				Error:  fmt.Sprintf("malformed env.var GITHUB_REF: expected to match regex '^refs/tags/(?:a)(?P<version>%v)$', got 'refs/tags/av1.0.0'", changelog.SemVerRegex),
			},
		},
		"Invalid Semver": {
			GitHubRef:     "refs/tags/1",
			GitHubSha:     "111",
			UnreleasedTag: "",
			Prefix:        "",
			Unreleased:    false,
			Expected: expected{
				Result: nil,
				Error:  fmt.Sprintf("malformed env.var GITHUB_REF: expected to match regex '^refs/tags/(?:v?)(?P<version>%v)$', got 'refs/tags/1'", changelog.SemVerRegex),
			},
		},
		"Complex Semver": {
			GitHubRef:     "refs/tags/v1.0.0-alpha-a.b-c-somethinglong+build.1-aef.1-its-okay",
			GitHubSha:     "111",
			UnreleasedTag: "",
			Prefix:        "",
			Unreleased:    false,
			Expected: expected{
				Result: &release.Reference{
					CommitHash: "111",
					Version:    "1.0.0-alpha-a.b-c-somethinglong+build.1-aef.1-its-okay",
					Tag:        "v1.0.0-alpha-a.b-c-somethinglong+build.1-aef.1-its-okay",
					Major:      "1",
					Minor:      "0",
					Patch:      "0",
					Prerelease: "alpha-a.b-c-somethinglong",
				},
				Error: "",
			},
		},
		"Complex Semver with Custom Prefix": {
			GitHubRef:     "refs/tags/1.0.01.0.0-alpha-a.b-c-somethinglong+build.1-aef.1-its-okay",
			GitHubSha:     "111",
			UnreleasedTag: "",
			Prefix:        "1.0.0",
			Unreleased:    false,
			Expected: expected{
				Result: &release.Reference{
					CommitHash: "111",
					Version:    "1.0.0-alpha-a.b-c-somethinglong+build.1-aef.1-its-okay",
					Tag:        "1.0.01.0.0-alpha-a.b-c-somethinglong+build.1-aef.1-its-okay",
					Major:      "1",
					Minor:      "0",
					Patch:      "0",
					Prerelease: "alpha-a.b-c-somethinglong",
				},
				Error: "",
			},
		},
		"Triggering Loop": {
			GitHubRef:     "refs/tags/" + release.UnreleasedDefaultTag,
			GitHubSha:     "111",
			UnreleasedTag: "",
			Prefix:        "",
			Unreleased:    true,
			Expected: expected{
				Result: nil,
				Error:  "workflow configuration error detected: trigger loop (triggering tag will be recreated and trigger the workflow again)",
			},
		},
		"Unreleased": {
			GitHubRef:     "refs/heads/master",
			GitHubSha:     "111",
			UnreleasedTag: "",
			Prefix:        "",
			Unreleased:    true,
			Expected: expected{
				Result: &release.Reference{
					CommitHash: "111",
					Version:    "Unreleased",
					Tag:        release.UnreleasedDefaultTag,
				},
				Error: "",
			},
		},
		"Unreleased with custom Tag": {
			GitHubRef:     "refs/heads/master",
			GitHubSha:     "111",
			UnreleasedTag: "future",
			Prefix:        "",
			Unreleased:    true,
			Expected: expected{
				Result: &release.Reference{
					CommitHash: "111",
					Version:    "Unreleased",
					Tag:        "future",
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
		if err := os.Setenv("UNRELEASED_TAG", test.UnreleasedTag); err != nil {
			t.Errorf("error preparing test case: error setting environmental variable UNRELEASED_TAG=%v: %v", test.UnreleasedTag, err)
			continue
		}
		time.Sleep(30 * time.Millisecond)

		// test
		r, err := release.GetReference(test.Prefix, test.Unreleased)
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
		if err := os.Unsetenv("UNRELEASED_TAG"); err != nil {
			t.Errorf("error cleanup: error unsetting environmental variable UNRELEASED_TAG: %v", err)
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
		Unreleased       bool
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
			Unreleased:       false,
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
						Major:      "1",
						Minor:      "0",
						Patch:      "0",
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
			Unreleased:       false,
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
						Major:      "1",
						Minor:      "0",
						Patch:      "0",
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
			Unreleased:       false,
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
						Major:      "1",
						Minor:      "0",
						Patch:      "0",
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
			Unreleased:       false,
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
						Major:      "1",
						Minor:      "0",
						Patch:      "0",
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
			Unreleased:       false,
			Files:            []string{},
			Expected: expected{
				Result: nil,
				Error:  fmt.Sprintf("error retrieving source code reference (control tag prefix via env.var TAG_PREFIX_REGEX): malformed env.var GITHUB_REF: expected to match regex '^refs/tags/(?:v?)(?P<version>%v)$', got 'refs/tags/1'", changelog.SemVerRegex),
			},
		},
		"Invalid Slug": {
			GitHubRef:        "refs/tags/1.0.0",
			GitHubSha:        "111",
			GitHubRepository: "anton-yurchenkogit-release",
			TagPrefix:        "",
			DraftRelease:     "false",
			PreRelease:       "false",
			Unreleased:       false,
			Files:            []string{},
			Expected: expected{
				Result: nil,
				Error:  fmt.Sprintf("error retrieving repository slug: malformed GITHUB_REPOSITORY (expected '%v', received 'anton-yurchenkogit-release')", release.SlugRegex),
			},
		},
		"Unreleased": {
			GitHubRef:        "refs/heads/master",
			GitHubSha:        "111",
			GitHubRepository: "anton-yurchenko/git-release",
			TagPrefix:        "",
			DraftRelease:     "false",
			PreRelease:       "false",
			Unreleased:       true,
			Files:            []string{"file1", "file2"},
			Expected: expected{
				Result: &release.Release{
					Name: "Latest",
					Slug: &release.Slug{
						Owner: "anton-yurchenko",
						Name:  "git-release",
					},
					Reference: &release.Reference{
						CommitHash: "111",
						Tag:        release.UnreleasedDefaultTag,
						Version:    "Unreleased",
					},
					Draft:      false,
					PreRelease: true,
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
		"Unreleased with Pre Release tag": {
			GitHubRef:        "refs/heads/master",
			GitHubSha:        "111",
			GitHubRepository: "anton-yurchenko/git-release",
			TagPrefix:        "",
			DraftRelease:     "false",
			PreRelease:       "true",
			Unreleased:       true,
			Files:            []string{},
			Expected: expected{
				Result: &release.Release{
					Name: "Latest",
					Slug: &release.Slug{
						Owner: "anton-yurchenko",
						Name:  "git-release",
					},
					Reference: &release.Reference{
						CommitHash: "111",
						Tag:        release.UnreleasedDefaultTag,
						Version:    "Unreleased",
					},
					Draft:      false,
					PreRelease: true,
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
		r, err := release.GetRelease(fs, test.Files, test.TagPrefix, test.Unreleased)
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
	// Publish uploads assets, and a mocked upload failure otherwise pays the real
	// 9 + 27 + 81 second backoff.
	defer release.SetRetryDelay(time.Millisecond)()

	a := assert.New(t)
	log.SetOutput(io.Discard)
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
		FailedAttempts         int
	}

	suite := map[string]test{
		// Every other fixture has Name == Tag and Draft == PreRelease == false, so
		// the mock argument matcher cannot tell those payload fields apart -
		// swapping Draft with Prerelease, or Name with TagName, passed. These two
		// cases make every field distinguishable.
		"Draft Release With A Distinct Title": {
			Release: &release.Release{
				Name: "Ship It",
				Slug: &release.Slug{
					Owner: "anton-yurchenko",
					Name:  "git-release",
				},
				Reference: &release.Reference{
					CommitHash: "111",
					Tag:        "v1.0.0",
					Version:    "1.0.0",
				},
				Draft:      true,
				PreRelease: false,
				Assets:     nil,
				Body:       "changelog",
			},
			CreateReleaseMock: createReleaseMock{
				Output: nil,
				Error:  nil,
			},
			UploadReleaseAssetMock: []error{},
			ExpectedError:          "",
		},
		"Pre Release With A Distinct Title": {
			Release: &release.Release{
				Name: "Nightly",
				Slug: &release.Slug{
					Owner: "anton-yurchenko",
					Name:  "git-release",
				},
				Reference: &release.Reference{
					CommitHash: "111",
					Tag:        "v1.0.0-rc.1",
					Version:    "1.0.0-rc.1",
				},
				Draft:      false,
				PreRelease: true,
				Assets:     nil,
				Body:       "changelog",
			},
			CreateReleaseMock: createReleaseMock{
				Output: nil,
				Error:  nil,
			},
			UploadReleaseAssetMock: []error{},
			ExpectedError:          "",
		},
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
					Major:      "1",
					Minor:      "0",
					Patch:      "0",
				},
				Draft:      false,
				PreRelease: false,
				Assets:     nil,
				Body:       "changelog",
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
					Major:      "1",
					Minor:      "0",
					Patch:      "0",
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
				Body: "changelog",
			},
			CreateReleaseMock: createReleaseMock{
				Output: &github.RepositoryRelease{
					ID: int64P(2),
				},
				Error: nil,
			},
			UploadReleaseAssetMock: []error{nil, nil},
			ExpectedError:          "",
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
					Major:      "1",
					Minor:      "0",
					Patch:      "0",
				},
				Draft:      false,
				PreRelease: false,
				Assets:     nil,
				Body:       "changelog",
			},
			CreateReleaseMock: createReleaseMock{
				Output: nil,
				Error:  errors.New("reason"),
			},
			ExpectedError: "reason",
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
					Major:      "1",
					Minor:      "0",
					Patch:      "0",
				},
				Draft:      false,
				PreRelease: false,
				Assets: &[]release.Asset{
					{
						Name: "file1",
						Path: "file1",
					},
				},
				Body: "changelog",
			},
			CreateReleaseMock: createReleaseMock{
				Output: &github.RepositoryRelease{
					ID: int64P(2),
				},
				Error: nil,
			},
			UploadReleaseAssetMock: []error{errors.New("reason")},
			ExpectedError:          "error uploading assets",
		},
	}

	var counter int
main:
	for name, test := range suite {
		counter++
		t.Logf("Test Case %v/%v - %s", counter, len(suite), name)

		// prepare test case
		//
		// NOTE: the file is created for EVERY asset, including those whose upload
		// is mocked to fail. Skipping it made Upload short-circuit in os.Open, so
		// the mocked API error was never reached and the case passed only because
		// Publish collapses every asset failure into the same message.
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
		m := mocks.NewRepositoriesClient(t)

		m.On("CreateRelease",
			context.Background(),
			test.Release.Slug.Owner,
			test.Release.Slug.Name,
			&github.RepositoryRelease{
				Name:            &test.Release.Name,
				TagName:         &test.Release.Reference.Tag,
				TargetCommitish: &test.Release.Reference.CommitHash,
				Body:            &test.Release.Body,
				Draft:           &test.Release.Draft,
				Prerelease:      &test.Release.PreRelease,
			}).Return(test.CreateReleaseMock.Output, nil, test.CreateReleaseMock.Error).Once()

		if test.Release.Assets != nil {
			for i, asset := range *test.Release.Assets {
				// A failing upload is retried until the attempts are exhausted;
				// a successful one is called exactly once.
				attempts := 1
				if test.UploadReleaseAssetMock[i] != nil {
					attempts = 4
				}

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
					mock.AnythingOfType("*os.File")).Return(nil, nil, test.UploadReleaseAssetMock[i]).Times(attempts)
			}
		}

		err := test.Release.Publish(m)
		if test.ExpectedError != "" || err != nil {
			a.EqualError(err, test.ExpectedError)
		}

		// cleanup
		if test.Release.Assets != nil {
			for i, asset := range *test.Release.Assets {
				if test.UploadReleaseAssetMock[i] == nil {
					if err := fs.Remove(asset.Path); err != nil {
						t.Errorf("error cleanup: error removing file %v: %v", asset.Path, err)
						continue main
					}
				}
			}
		}
		time.Sleep(30 * time.Millisecond)
	}
}

func TestDeleteUnreleased(t *testing.T) {
	a := assert.New(t)
	log.SetOutput(io.Discard)

	type getReleaseByTagMock struct {
		Output *github.RepositoryRelease
		Error  error
	}

	type test struct {
		Release                *release.Release
		GetReleaseByTagMock    getReleaseByTagMock
		DeleteReleaseMockError error
		DeleteRefMockError     error
		SkipDeleteRefMock      bool
		GetRefMockErrors       []error
		ExpectedError          string
	}

	suite := map[string]test{
		"Success": {
			Release: &release.Release{
				Name: "Latest",
				Slug: &release.Slug{
					Owner: "anton-yurchenko",
					Name:  "git-release",
				},
				Reference: &release.Reference{
					CommitHash: "111",
					Tag:        "latest",
					Version:    "Unrelease",
				},
				Draft:      false,
				PreRelease: false,
				Assets:     nil,
				Body:       "changelog",
			},
			GetReleaseByTagMock: getReleaseByTagMock{
				Output: &github.RepositoryRelease{
					ID:   int64P(1),
					Name: stringP("Latest"),
				},
				Error: nil,
			},
			DeleteReleaseMockError: nil,
			DeleteRefMockError:     nil,
			GetRefMockErrors: []error{
				nil,
				errors.New("404 Not Found"),
			},
			ExpectedError: "",
		},
		"GetReleaseByTag error": {
			Release: &release.Release{
				Name: "Latest",
				Slug: &release.Slug{
					Owner: "anton-yurchenko",
					Name:  "git-release",
				},
				Reference: &release.Reference{
					CommitHash: "111",
					Tag:        "latest",
					Version:    "Unrelease",
				},
				Draft:      false,
				PreRelease: false,
				Assets:     nil,
				Body:       "changelog",
			},
			GetReleaseByTagMock: getReleaseByTagMock{
				Output: nil,
				Error:  errors.New("reason"),
			},
			DeleteReleaseMockError: nil,
			DeleteRefMockError:     nil,
			GetRefMockErrors:       []error{},
			ExpectedError:          "error retrieving a precedent release with a tag latest: reason",
		},
		"DeleteRelease error": {
			Release: &release.Release{
				Name: "Latest",
				Slug: &release.Slug{
					Owner: "anton-yurchenko",
					Name:  "git-release",
				},
				Reference: &release.Reference{
					CommitHash: "111",
					Tag:        "latest",
					Version:    "Unrelease",
				},
				Draft:      false,
				PreRelease: false,
				Assets:     nil,
				Body:       "changelog",
			},
			GetReleaseByTagMock: getReleaseByTagMock{
				Output: &github.RepositoryRelease{
					ID:   int64P(1),
					Name: stringP("Latest"),
				},
				Error: nil,
			},
			DeleteReleaseMockError: errors.New("reason"),
			// NOTE: no DeleteRef expectation - the run aborts on the release
			// deletion, so registering one would go unmet.
			SkipDeleteRefMock: true,
			GetRefMockErrors:  []error{},
			ExpectedError:     "error deleting precedent release: reason",
		},
		"DeleteRef error": {
			Release: &release.Release{
				Name: "Latest",
				Slug: &release.Slug{
					Owner: "anton-yurchenko",
					Name:  "git-release",
				},
				Reference: &release.Reference{
					CommitHash: "111",
					Tag:        "latest",
					Version:    "Unrelease",
				},
				Draft:      false,
				PreRelease: false,
				Assets:     nil,
				Body:       "changelog",
			},
			GetReleaseByTagMock: getReleaseByTagMock{
				Output: &github.RepositoryRelease{
					ID:   int64P(1),
					Name: stringP("Latest"),
				},
				Error: nil,
			},
			DeleteReleaseMockError: nil,
			DeleteRefMockError:     errors.New("reason"),
			GetRefMockErrors:       []error{},
			ExpectedError:          "error deleting precedent tag: reason",
		},
		"GetRef error": {
			Release: &release.Release{
				Name: "Latest",
				Slug: &release.Slug{
					Owner: "anton-yurchenko",
					Name:  "git-release",
				},
				Reference: &release.Reference{
					CommitHash: "111",
					Tag:        "latest",
					Version:    "Unrelease",
				},
				Draft:      false,
				PreRelease: false,
				Assets:     nil,
				Body:       "changelog",
			},
			GetReleaseByTagMock: getReleaseByTagMock{
				Output: &github.RepositoryRelease{
					ID:   int64P(1),
					Name: stringP("Latest"),
				},
				Error: nil,
			},
			DeleteReleaseMockError: nil,
			DeleteRefMockError:     nil,
			GetRefMockErrors:       []error{errors.New("reason")},
			ExpectedError:          "error fetching precedent tag: reason",
		},
	}

	var counter int
	for name, test := range suite {
		counter++
		t.Logf("Test Case %v/%v - %s", counter, len(suite), name)

		// test
		tag := fmt.Sprintf("refs/tags/%v", test.Release.Reference.Tag)
		repoMock := mocks.NewRepositoriesClient(t)
		gitMock := mocks.NewGitClient(t)

		repoMock.On("GetReleaseByTag",
			context.Background(),
			test.Release.Slug.Owner,
			test.Release.Slug.Name,
			test.Release.Reference.Tag).Return(test.GetReleaseByTagMock.Output, nil, test.GetReleaseByTagMock.Error).Once()

		if test.GetReleaseByTagMock.Output != nil {
			repoMock.On("DeleteRelease",
				context.Background(),
				test.Release.Slug.Owner,
				test.Release.Slug.Name,
				*test.GetReleaseByTagMock.Output.ID).Return(nil, test.DeleteReleaseMockError).Once()

			if !test.SkipDeleteRefMock {
				gitMock.On("DeleteRef",
					context.Background(),
					test.Release.Slug.Owner,
					test.Release.Slug.Name,
					tag).Return(nil, test.DeleteRefMockError).Once()
			}

			for _, e := range test.GetRefMockErrors {
				gitMock.On("GetRef",
					context.Background(),
					test.Release.Slug.Owner,
					test.Release.Slug.Name,
					tag).Return(nil, nil, e).Once()
			}
		}

		err := test.Release.DeleteUnreleased(repoMock, gitMock)
		if test.ExpectedError != "" || err != nil {
			a.EqualError(err, test.ExpectedError)
		}
	}
}

func TestUpdateUnreleasedTag(t *testing.T) {
	a := assert.New(t)

	type test struct {
		Release            *release.Release
		CreateRefMockError error
		ExpectedError      string
	}

	suite := map[string]test{
		"Success": {
			Release: &release.Release{
				Name: "Latest",
				Slug: &release.Slug{
					Owner: "anton-yurchenko",
					Name:  "git-release",
				},
				Reference: &release.Reference{
					CommitHash: "111",
					Tag:        "latest",
					Version:    "Unrelease",
				},
				Draft:      false,
				PreRelease: false,
				Assets:     nil,
				Body:       "changelog",
			},
			CreateRefMockError: nil,
			ExpectedError:      "",
		},
		"Error": {
			Release: &release.Release{
				Name: "Latest",
				Slug: &release.Slug{
					Owner: "anton-yurchenko",
					Name:  "git-release",
				},
				Reference: &release.Reference{
					CommitHash: "111",
					Tag:        "latest",
					Version:    "Unrelease",
				},
				Draft:      false,
				PreRelease: false,
				Assets:     nil,
				Body:       "changelog",
			},
			CreateRefMockError: errors.New("reason"),
			ExpectedError:      "reason",
		},
	}

	var counter int
	for name, test := range suite {
		counter++
		t.Logf("Test Case %v/%v - %s", counter, len(suite), name)

		// test
		tag := fmt.Sprintf("refs/tags/%v", test.Release.Reference.Tag)
		gitMock := mocks.NewGitClient(t)

		gitMock.On("CreateRef",
			context.Background(),
			test.Release.Slug.Owner,
			test.Release.Slug.Name,
			github.CreateRef{
				Ref: tag,
				SHA: test.Release.Reference.CommitHash,
			}).Return(nil, nil, test.CreateRefMockError).Once()

		err := test.Release.UpdateUnreleasedTag(gitMock)
		if test.ExpectedError != "" || err != nil {
			a.EqualError(err, test.ExpectedError)
		}
	}
}

// TestGetReleaseConfigurationErrors covers the failure paths that reject a
// malformed configuration before anything is created.
func TestGetReleaseConfigurationErrors(t *testing.T) {
	a := assert.New(t)
	log.SetOutput(io.Discard)

	type test struct {
		Env           map[string]string
		TagPrefix     string
		Unreleased    bool
		ExpectedError string
	}

	suite := map[string]test{
		// v6 accepted any value and silently treated it as false, so a user
		// asking for a draft got a public release.
		"Invalid DRAFT_RELEASE": {
			Env:           map[string]string{"DRAFT_RELEASE": "yes"},
			ExpectedError: "DRAFT_RELEASE expects 'true' or 'false', received 'yes'",
		},
		"Invalid PRE_RELEASE": {
			Env:           map[string]string{"PRE_RELEASE": "1"},
			ExpectedError: "PRE_RELEASE expects 'true' or 'false', received '1'",
		},
		// v6 used regexp.MustCompile on this value, so a configuration mistake
		// terminated the action with a panic (exit 2) instead of an error.
		"Malformed TAG_PREFIX_REGEX": {
			TagPrefix:     "[a-",
			ExpectedError: "malformed env.var TAG_PREFIX_REGEX",
		},
	}

	var counter int
	for name, test := range suite {
		counter++
		t.Logf("Test Case %v/%v - %s", counter, len(suite), name)

		t.Setenv("GITHUB_REF", "refs/tags/v1.0.0")
		t.Setenv("GITHUB_SHA", "111")
		t.Setenv("GITHUB_REPOSITORY", "anton-yurchenko/git-release")
		t.Setenv("DRAFT_RELEASE", "")
		t.Setenv("PRE_RELEASE", "")

		for k, v := range test.Env {
			t.Setenv(k, v)
		}

		r, err := release.GetRelease(afero.NewMemMapFs(), nil, test.TagPrefix, test.Unreleased)
		a.Nil(r)
		a.ErrorContains(err, test.ExpectedError)
	}
}

// TestDeleteUnreleasedTolerance pins the branches that treat a missing release
// or tag as success rather than as an error.
//
// NOTE: these paths classify GitHub errors by SUBSTRING on err.Error(). That is
// fragile - a change to go-github's error format would silently turn "nothing to
// clean up" into a hard failure - so the exact strings are asserted here.
func TestDeleteUnreleasedTolerance(t *testing.T) {
	a := assert.New(t)
	log.SetOutput(io.Discard)

	rel := &release.Release{
		Name:      "Latest",
		Slug:      &release.Slug{Owner: "anton-yurchenko", Name: "git-release"},
		Reference: &release.Reference{CommitHash: "111", Tag: "latest", Version: release.UnreleasedVersion},
	}

	t.Run("no precedent release and no precedent tag", func(t *testing.T) {
		repoMock := mocks.NewRepositoriesClient(t)
		gitMock := mocks.NewGitClient(t)

		repoMock.On("GetReleaseByTag", context.Background(), "anton-yurchenko", "git-release", "latest").
			Return(nil, nil, errors.New("GET https://api.github.com/...: 404 Not Found []")).Once()
		gitMock.On("DeleteRef", context.Background(), "anton-yurchenko", "git-release", "refs/tags/latest").
			Return(nil, errors.New("DELETE https://api.github.com/...: 422 Reference does not exist []")).Once()

		a.NoError(rel.DeleteUnreleased(repoMock, gitMock))
		repoMock.AssertExpectations(t)
		gitMock.AssertExpectations(t)
	})

	// Any error that is NOT the tolerated one must abort.
	t.Run("an unexpected release error aborts", func(t *testing.T) {
		repoMock := mocks.NewRepositoriesClient(t)
		gitMock := mocks.NewGitClient(t)

		repoMock.On("GetReleaseByTag", context.Background(), "anton-yurchenko", "git-release", "latest").
			Return(nil, nil, errors.New("GET https://api.github.com/...: 500 Internal Server Error []")).Once()

		err := rel.DeleteUnreleased(repoMock, gitMock)
		a.ErrorContains(err, "error retrieving a precedent release with a tag latest")
		repoMock.AssertExpectations(t)
	})

	t.Run("an unexpected tag error aborts", func(t *testing.T) {
		repoMock := mocks.NewRepositoriesClient(t)
		gitMock := mocks.NewGitClient(t)

		repoMock.On("GetReleaseByTag", context.Background(), "anton-yurchenko", "git-release", "latest").
			Return(nil, nil, errors.New("404 Not Found")).Once()
		gitMock.On("DeleteRef", context.Background(), "anton-yurchenko", "git-release", "refs/tags/latest").
			Return(nil, errors.New("403 Forbidden")).Once()

		err := rel.DeleteUnreleased(repoMock, gitMock)
		a.ErrorContains(err, "error deleting precedent tag")
		repoMock.AssertExpectations(t)
		gitMock.AssertExpectations(t)
	})
}
