package release_test

import (
	"context"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/anton-yurchenko/git-release/mocks"
	"github.com/anton-yurchenko/git-release/release"
	"github.com/google/go-github/github"
	"github.com/pkg/errors"
	"github.com/spf13/afero"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

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
	a := assert.New(t)
	fs := afero.NewOsFs()
	id := int64(1)

	type expected struct {
		Message string
		Error   string
	}

	type test struct {
		Asset    release.Asset
		Release  *release.Release
		Expected expected
		Error    error
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
			},
			Expected: expected{
				Message: "uploading asset testFile1",
				Error:   "",
			},
			Error: nil,
		},
		"Error": {
			Asset: release.Asset{
				Name: "testFile2",
				Path: "testFile2",
			},
			Release: &release.Release{
				Slug: &release.Slug{
					Owner: "anton-yurchenko",
					Name:  "git-release",
				},
			},
			Expected: expected{
				Message: "uploading asset testFile2",
				Error:   "reason",
			},
			Error: errors.New("reason"),
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
				Message: "uploading asset testFile3",
				Error:   "open testFile3: no such file or directory",
			},
			Error: nil,
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
		msgs := make(chan string, 1)
		errs := make(chan error, 1)

		m := new(mocks.RepositoriesClient)
		m.On("UploadReleaseAsset",
			context.Background(),
			test.Release.Slug.Owner,
			test.Release.Slug.Name,
			id,
			&github.UploadOptions{
				Name: strings.ReplaceAll(test.Asset.Name, "/", "-"),
			},
			mock.AnythingOfType("*os.File")).Return(nil, nil, test.Error).Once()

		test.Asset.Upload(test.Release, m, id, msgs, errs, wg)

		msg := <-msgs
		a.Equal(test.Expected.Message, msg)

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
