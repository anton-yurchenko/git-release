package release_test

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"testing"
	"time"

	"git-release/mocks"

	"git-release/release"

	"github.com/google/go-github/v78/github"
	"github.com/pkg/errors"
	"github.com/spf13/afero"

	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func pInt64(v int64) *int64 {
	return &v
}

func pString(v string) *string {
	return &v
}

func TestGetAssets(t *testing.T) {
	a := assert.New(t)
	roBase := afero.NewReadOnlyFs(afero.NewOsFs())
	fs := afero.NewCopyOnWriteFs(roBase, afero.NewMemMapFs())

	type expected struct {
		Result *[]release.Asset
		Error  string
	}

	type test struct {
		Patterns []string
		Files    []string
		Expected expected
	}

	// NOTE: patterns arrive already split, one per line - the separator guessing
	// that used to live here moved out of the asset list entirely, and is
	// covered by TestGetAssets in the main package.
	suite := map[string]test{
		"No Patterns": {
			Patterns: []string{},
			Files:    []string{},
			Expected: expected{
				Result: &[]release.Asset{},
			},
		},
		"Single File": {
			Patterns: []string{"file1"},
			Files:    []string{"file1"},
			Expected: expected{
				Result: &[]release.Asset{
					{Name: "file1", Path: "file1"},
				},
			},
		},
		"Multiple Files": {
			Patterns: []string{"file1", "file2"},
			Files:    []string{"file1", "file2"},
			Expected: expected{
				Result: &[]release.Asset{
					{Name: "file1", Path: "file1"},
					{Name: "file2", Path: "file2"},
				},
			},
		},
		"Glob": {
			Patterns: []string{"*.zip"},
			Files:    []string{"a.zip", "b.zip", "c.txt"},
			Expected: expected{
				Result: &[]release.Asset{
					{Name: "a.zip", Path: "a.zip"},
					{Name: "b.zip", Path: "b.zip"},
				},
			},
		},
		"Path Containing A Space": {
			Patterns: []string{"release notes.txt"},
			Files:    []string{"release notes.txt"},
			Expected: expected{
				Result: &[]release.Asset{
					{Name: "release notes.txt", Path: "release notes.txt"},
				},
			},
		},
		"Path Containing A Comma": {
			Patterns: []string{"a,b.zip"},
			Files:    []string{"a,b.zip"},
			Expected: expected{
				Result: &[]release.Asset{
					{Name: "a,b.zip", Path: "a,b.zip"},
				},
			},
		},
		"Unmatched Pattern": {
			Patterns: []string{"nothing-*.zip"},
			Files:    []string{},
			Expected: expected{
				Result: &[]release.Asset{},
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

		// test
		r, err := release.GetAssets(fs, test.Patterns)
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
	}
}

func TestUpload(t *testing.T) {
	log.SetOutput(io.Discard)

	// The retry LOGIC is what matters here, not 117 seconds of real sleeping.
	defer release.SetRetryDelay(time.Millisecond)()

	a := assert.New(t)
	fs := afero.NewOsFs()
	id := int64(1)

	type expected struct {
		Message string
		Error   string
	}

	type mockResponses struct {
		LastTry                    bool
		UploadReleaseAssetResponse *github.Response
		UploadReleaseAssetError    error
		GetReleaseByTagRelease     *github.RepositoryRelease
		GetReleaseByTagError       error
		DeleteReleaseAssetError    error
	}

	type test struct {
		Asset         release.Asset
		Release       *release.Release
		MockResponses []mockResponses
		Expected      expected
	}

	suite := map[string]test{
		"Success": {
			Asset: release.Asset{
				Name: "testFile1",
				Path: "testFile1",
			},
			Release: &release.Release{
				Slug: &release.Slug{
					Owner: "anton-yurchenko",
					Name:  "git-release",
				},
				Reference: &release.Reference{
					Tag: "v1.0.0",
				},
			},
			MockResponses: []mockResponses{
				{
					UploadReleaseAssetResponse: &github.Response{
						Response: &http.Response{StatusCode: http.StatusOK},
					},
					UploadReleaseAssetError: nil,
				},
			},
			Expected: expected{
				Error: "",
			},
		},
		"Ghost Release Asset Not Found - Recovers On Retry": {
			Asset: release.Asset{
				Name: "testFile1",
				Path: "testFile1",
			},
			Release: &release.Release{
				Slug: &release.Slug{
					Owner: "anton-yurchenko",
					Name:  "git-release",
				},
				Reference: &release.Reference{
					Tag: "v1.0.0",
				},
			},
			MockResponses: []mockResponses{
				{
					UploadReleaseAssetResponse: &github.Response{
						Response: &http.Response{StatusCode: http.StatusBadGateway},
					},
					UploadReleaseAssetError: errors.New("reason-c"),
					GetReleaseByTagRelease: &github.RepositoryRelease{
						Assets: []*github.ReleaseAsset{
							{
								ID:   pInt64(123),
								Name: pString("testFile2"),
							},
						},
					},
					GetReleaseByTagError: nil,
				},
				{
					LastTry: true,
					UploadReleaseAssetResponse: &github.Response{
						Response: &http.Response{StatusCode: http.StatusOK},
					},
					UploadReleaseAssetError: nil,
				},
			},
			Expected: expected{
				// The ghost is NOT found on the first attempt, so uploadHandler
				// reports it, Upload treats that as retryable, and the second
				// attempt succeeds. The run therefore returns no error.
				Error: "",
			},
		},
		"Asset Already Exists - Last Try": {
			Asset: release.Asset{
				Name: "test/File1",
				Path: "testFile1",
			},
			Release: &release.Release{
				Slug: &release.Slug{
					Owner: "anton-yurchenko",
					Name:  "git-release",
				},
				Reference: &release.Reference{
					Tag: "v1.0.0",
				},
			},
			MockResponses: []mockResponses{
				{
					UploadReleaseAssetResponse: &github.Response{
						Response: &http.Response{StatusCode: http.StatusInternalServerError},
					},
					UploadReleaseAssetError: errors.New("reason-a"),
				},
				{
					UploadReleaseAssetResponse: &github.Response{
						Response: &http.Response{StatusCode: http.StatusBadGateway},
					},
					UploadReleaseAssetError: errors.New("reason-c"),
					GetReleaseByTagRelease: &github.RepositoryRelease{
						Assets: []*github.ReleaseAsset{
							{
								ID:   pInt64(123),
								Name: pString("test-File1"),
							},
						},
					},
					GetReleaseByTagError: errors.New("reason-d"),
				},
				{
					UploadReleaseAssetResponse: &github.Response{
						Response: &http.Response{StatusCode: http.StatusUnprocessableEntity},
					},
					UploadReleaseAssetError: errors.New("reason-c"),
					GetReleaseByTagRelease: &github.RepositoryRelease{
						Assets: []*github.ReleaseAsset{
							{
								ID:   pInt64(123),
								Name: pString("test-File1"),
							},
						},
					},
					GetReleaseByTagError:    nil,
					DeleteReleaseAssetError: errors.New("reason"),
				},
				{
					LastTry: true,
					UploadReleaseAssetResponse: &github.Response{
						Response: &http.Response{StatusCode: http.StatusBadGateway},
					},
					UploadReleaseAssetError: errors.New("reason-e"),
				},
			},
			Expected: expected{
				Error: "maximum attempts reached uploading asset: test/File1",
			},
		},
		"Recover": {
			Asset: release.Asset{
				Name: "test/File1",
				Path: "testFile1",
			},
			Release: &release.Release{
				Slug: &release.Slug{
					Owner: "anton-yurchenko",
					Name:  "git-release",
				},
				Reference: &release.Reference{
					Tag: "v1.0.0",
				},
			},
			MockResponses: []mockResponses{
				{
					UploadReleaseAssetResponse: &github.Response{
						Response: &http.Response{StatusCode: http.StatusInternalServerError},
					},
					UploadReleaseAssetError: errors.New("reason-a"),
				},
				{
					UploadReleaseAssetResponse: &github.Response{
						Response: &http.Response{StatusCode: http.StatusUnprocessableEntity},
					},
					UploadReleaseAssetError: errors.New("reason-b"),
					GetReleaseByTagRelease: &github.RepositoryRelease{
						Assets: []*github.ReleaseAsset{
							{
								ID:   pInt64(123),
								Name: pString("test-File1"),
							},
						},
					},
					GetReleaseByTagError:    nil,
					DeleteReleaseAssetError: nil,
				},
				{
					LastTry: true,
					UploadReleaseAssetResponse: &github.Response{
						Response: &http.Response{StatusCode: http.StatusOK},
					},
					UploadReleaseAssetError: nil,
				},
			},
			Expected: expected{
				Error: "",
			},
		},
		"No API Response": {
			Asset: release.Asset{
				Name: "test/File1",
				Path: "testFile1",
			},
			Release: &release.Release{
				Slug: &release.Slug{
					Owner: "anton-yurchenko",
					Name:  "git-release",
				},
				Reference: &release.Reference{
					Tag: "v1.0.0",
				},
			},
			MockResponses: []mockResponses{
				{
					UploadReleaseAssetResponse: &github.Response{
						Response: nil,
					},
					UploadReleaseAssetError: errors.New("reason-a"),
				},
				{
					LastTry: true,
					UploadReleaseAssetResponse: &github.Response{
						Response: &http.Response{StatusCode: http.StatusOK},
					},
					UploadReleaseAssetError: nil,
				},
			},
			Expected: expected{
				Error: "",
			},
		},
		"File Does Not Exists": {
			Asset: release.Asset{
				Name: "testFile3",
				Path: "testFile3",
			},
			Release: &release.Release{
				Slug: &release.Slug{
					Owner: "anton-yurchenko",
					Name:  "git-release",
				},
			},
			Expected: expected{
				Error: "error opening a file: open testFile3: no such file or directory",
			},
		},
	}

	var counter int
	for name, test := range suite {
		counter++
		t.Logf("Test Case %v/%v - %s", counter, len(suite), name)

		// prepare test case
		if name != "File Does Not Exists" {
			if err := afero.WriteFile(fs, test.Asset.Path, []byte(""), 0644); err != nil {
				t.Errorf("error preparing test case: error creating file %v: %v", test.Asset.Path, err)
				continue
			}
		}
		time.Sleep(30 * time.Millisecond)

		// test
		wg := new(sync.WaitGroup)
		wg.Add(1)
		errs := make(chan error, 1)

		m := mocks.NewRepositoriesClient(t)
		for _, res := range test.MockResponses {
			m.On("UploadReleaseAsset",
				context.Background(),
				test.Release.Slug.Owner,
				test.Release.Slug.Name,
				id,
				&github.UploadOptions{
					Name: strings.ReplaceAll(test.Asset.Name, "/", "-"),
				},
				mock.AnythingOfType("*os.File"),
			).Return(nil, res.UploadReleaseAssetResponse, res.UploadReleaseAssetError).Once()

			if !res.LastTry && res.UploadReleaseAssetResponse.Response != nil {
				if res.UploadReleaseAssetResponse.StatusCode == http.StatusBadGateway || res.UploadReleaseAssetResponse.StatusCode == http.StatusUnprocessableEntity {
					m.On("GetReleaseByTag",
						context.Background(),
						test.Release.Slug.Owner,
						test.Release.Slug.Name,
						test.Release.Reference.Tag,
					).Return(res.GetReleaseByTagRelease, nil, res.GetReleaseByTagError).Once()

					if res.GetReleaseByTagError == nil {
						var assetID int64
						for _, s := range res.GetReleaseByTagRelease.Assets {
							if *s.Name == strings.ReplaceAll(test.Asset.Name, "/", "-") {
								assetID = *s.ID
								break
							}
						}

						if res.GetReleaseByTagError == nil && assetID != 0 {
							m.On("DeleteReleaseAsset",
								context.Background(),
								test.Release.Slug.Owner,
								test.Release.Slug.Name,
								assetID,
							).Return(nil, res.DeleteReleaseAssetError).Once()
						}
					}
				}
			}
		}

		test.Asset.Upload(test.Release, m, id, errs, wg)

		// NOTE: unconditional. Guarding this with `if err != nil` meant a case
		// whose expected error stopped occurring asserted nothing at all - four
		// of the six cases were silently checking nothing, and replacing the
		// "maximum attempts reached" error with nil passed the whole suite.
		err := <-errs
		if test.Expected.Error == "" {
			a.NoError(err)
		} else {
			a.EqualError(err, test.Expected.Error)
		}

		wg.Wait()

		// cleanup
		if name != "File Does Not Exists" {
			if err := fs.Remove(test.Asset.Path); err != nil {
				t.Errorf("error cleanup: error removing file %v: %v", test.Asset.Path, err)
			}
			time.Sleep(30 * time.Millisecond)
		}
	}
}

// TestUploadHandlerGhostRecovery pins the 502/422 recovery branch.
//
// GitHub can report a failed upload that nevertheless left the asset attached to
// the release. The next attempt would then fail with "already exists" forever, so
// the handler deletes that ghost and asks to be retried. Upload() reports only the
// final outcome, so this branch is only observable one attempt at a time.
func TestUploadHandlerGhostRecovery(t *testing.T) {
	log.SetOutput(io.Discard)

	a := assert.New(t)
	const id int64 = 1

	rel := &release.Release{
		Slug:      &release.Slug{Owner: "anton-yurchenko", Name: "git-release"},
		Reference: &release.Reference{Tag: "v1.0.0"},
	}

	fs := afero.NewOsFs()
	if err := afero.WriteFile(fs, "ghostFile", []byte(""), 0644); err != nil {
		t.Fatalf("error preparing test case: %v", err)
	}
	defer func() { _ = fs.Remove("ghostFile") }()

	asset := &release.Asset{Name: "ghostFile", Path: "ghostFile"}

	for _, status := range []int{http.StatusBadGateway, http.StatusUnprocessableEntity} {
		t.Run(fmt.Sprintf("ghost deleted on %v", status), func(t *testing.T) {
			m := mocks.NewRepositoriesClient(t)

			m.On("UploadReleaseAsset", context.Background(), "anton-yurchenko", "git-release", id,
				&github.UploadOptions{Name: "ghostFile"}, mock.AnythingOfType("*os.File")).
				Return(nil, &github.Response{Response: &http.Response{StatusCode: status}}, errors.New("reason")).Once()

			m.On("GetReleaseByTag", context.Background(), "anton-yurchenko", "git-release", "v1.0.0").
				Return(&github.RepositoryRelease{
					Assets: []*github.ReleaseAsset{{ID: int64P(7), Name: stringP("ghostFile")}},
				}, nil, nil).Once()

			// The ghost carrying the same name must be removed, or every later
			// attempt collides with it.
			m.On("DeleteReleaseAsset", context.Background(), "anton-yurchenko", "git-release", int64(7)).
				Return(nil, nil).Once()

			err := asset.UploadHandler(rel, m, id, false)
			a.EqualError(err, "ghost release asset deleted")
		})
	}

	t.Run("no ghost to delete", func(t *testing.T) {
		m := mocks.NewRepositoriesClient(t)

		m.On("UploadReleaseAsset", context.Background(), "anton-yurchenko", "git-release", id,
			&github.UploadOptions{Name: "ghostFile"}, mock.AnythingOfType("*os.File")).
			Return(nil, &github.Response{Response: &http.Response{StatusCode: http.StatusBadGateway}}, errors.New("reason")).Once()

		m.On("GetReleaseByTag", context.Background(), "anton-yurchenko", "git-release", "v1.0.0").
			Return(&github.RepositoryRelease{
				Assets: []*github.ReleaseAsset{{ID: int64P(7), Name: stringP("somethingElse")}},
			}, nil, nil).Once()

		err := asset.UploadHandler(rel, m, id, false)
		a.EqualError(err, "ghost release asset not found")
	})

	// On the final attempt there is no point recovering - the error is reported.
	t.Run("no recovery on the last attempt", func(t *testing.T) {
		m := mocks.NewRepositoriesClient(t)

		m.On("UploadReleaseAsset", context.Background(), "anton-yurchenko", "git-release", id,
			&github.UploadOptions{Name: "ghostFile"}, mock.AnythingOfType("*os.File")).
			Return(nil, &github.Response{Response: &http.Response{StatusCode: http.StatusBadGateway}}, errors.New("reason")).Once()

		err := asset.UploadHandler(rel, m, id, true)
		a.EqualError(err, "reason")
	})
}
