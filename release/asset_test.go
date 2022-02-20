package release_test

import (
	"context"
	"io/ioutil"
	"net/http"
	"strings"
	"sync"
	"testing"
	"time"

	"git-release/mocks"

	"git-release/release"

	"github.com/google/go-github/github"
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
		Args     []string
		Files    []string
		Expected expected
	}

	suite := map[string]test{
		"Empty Args": {
			Args:  []string{},
			Files: []string{},
			Expected: expected{
				Result: &[]release.Asset{},
				Error:  "",
			},
		},
		"Space Separator": {
			Args:  []string{"file1 file2"},
			Files: []string{"file1", "file2"},
			Expected: expected{
				Result: &[]release.Asset{
					{
						Name: "file1",
						Path: "file1",
					},
					{
						Name: "file2",
						Path: "file2",
					},
				},
				Error: "",
			},
		},
		"New Line Separator": {
			Args:  []string{"file1\nfile2"},
			Files: []string{"file1", "file2"},
			Expected: expected{
				Result: &[]release.Asset{
					{
						Name: "file1",
						Path: "file1",
					},
					{
						Name: "file2",
						Path: "file2",
					},
				},
				Error: "",
			},
		},
		"Comma Separator": {
			Args:  []string{"file1,file2"},
			Files: []string{"file1", "file2"},
			Expected: expected{
				Result: &[]release.Asset{
					{
						Name: "file1",
						Path: "file1",
					},
					{
						Name: "file2",
						Path: "file2",
					},
				},
				Error: "",
			},
		},
		"Pipe Separator": {
			Args:  []string{"file1|file2"},
			Files: []string{"file1", "file2"},
			Expected: expected{
				Result: &[]release.Asset{
					{
						Name: "file1",
						Path: "file1",
					},
					{
						Name: "file2",
						Path: "file2",
					},
				},
				Error: "",
			},
		},
		"Multiple Separators": {
			Args:  []string{"file1 file2\nfile3,file4|file5"},
			Files: []string{"file1", "file2", "file3", "file4", "file5"},
			Expected: expected{
				Result: &[]release.Asset{
					{
						Name: "file1",
						Path: "file1",
					},
				},
				Error: "",
			},
		},
		"Multiple Arguments": {
			Args:  []string{"file1", "file2"},
			Files: []string{"file1", "file2"},
			Expected: expected{
				Result: &[]release.Asset{
					{
						Name: "file1",
						Path: "file1",
					},
					{
						Name: "file2",
						Path: "file2",
					},
				},
				Error: "",
			},
		},
		"Multiple Arguments with Space Separator": {
			Args:  []string{"file1 file2", "file3 file4"},
			Files: []string{"file1", "file2", "file3", "file4"},
			Expected: expected{
				Result: &[]release.Asset{
					{
						Name: "file1",
						Path: "file1",
					},
					{
						Name: "file2",
						Path: "file2",
					},
					{
						Name: "file3",
						Path: "file3",
					},
					{
						Name: "file4",
						Path: "file4",
					},
				},
				Error: "",
			},
		},
		"Multiple Arguments with New Line Separator": {
			Args:  []string{"file1\nfile2", "file3\nfile4"},
			Files: []string{"file1", "file2", "file3", "file4"},
			Expected: expected{
				Result: &[]release.Asset{
					{
						Name: "file1",
						Path: "file1",
					},
					{
						Name: "file2",
						Path: "file2",
					},
					{
						Name: "file3",
						Path: "file3",
					},
					{
						Name: "file4",
						Path: "file4",
					},
				},
				Error: "",
			},
		},
		"Multiple Arguments with Comma Separator": {
			Args:  []string{"file1,file2", "file3,file4"},
			Files: []string{"file1", "file2", "file3", "file4"},
			Expected: expected{
				Result: &[]release.Asset{
					{
						Name: "file1",
						Path: "file1",
					},
					{
						Name: "file2",
						Path: "file2",
					},
					{
						Name: "file3",
						Path: "file3",
					},
					{
						Name: "file4",
						Path: "file4",
					},
				},
				Error: "",
			},
		},
		"Multiple Arguments with Pipe Separator": {
			Args:  []string{"file1|file2", "file3|file4"},
			Files: []string{"file1", "file2", "file3", "file4"},
			Expected: expected{
				Result: &[]release.Asset{
					{
						Name: "file1",
						Path: "file1",
					},
					{
						Name: "file2",
						Path: "file2",
					},
					{
						Name: "file3",
						Path: "file3",
					},
					{
						Name: "file4",
						Path: "file4",
					},
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

		// test
		r, err := release.GetAssets(fs, test.Args)
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
	log.SetOutput(ioutil.Discard)

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
		"Ghost Release Asset Not Found [very long test]": {
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
						Assets: []github.ReleaseAsset{
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
				Error: "ghost release asset not found",
			},
		},
		"Asset Already Exists - Last Try [very long test]": {
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
						Assets: []github.ReleaseAsset{
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
						Assets: []github.ReleaseAsset{
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
		"Recover [long test]": {
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
						Assets: []github.ReleaseAsset{
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
		"No API Response [long test]": {
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

		m := new(mocks.RepositoriesClient)
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

		err := <-errs
		if err != nil {
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
