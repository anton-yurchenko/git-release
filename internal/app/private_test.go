package app_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/anton-yurchenko/git-release/internal/app"
	"github.com/anton-yurchenko/git-release/internal/pkg/asset"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
)

func TestIsExists(t *testing.T) {
	assert := assert.New(t)
	fs := afero.NewMemMapFs()

	// TEST: file not exist
	expected := false
	t.Log("Test Case 1/2 - File Does Not Exist")

	result, err := app.IsExists("does-not-exist", fs)

	assert.Equal(nil, err)
	assert.Equal(expected, result)

	// TEST: file exist
	expected = true
	t.Log("Test Case 2/2 - File Exist")

	_, err = fs.Create("exist")
	assert.Equal(nil, err, "preparation: error creating test file 'exist'")

	result, err = app.IsExists("exist", fs)

	assert.Equal(nil, err)
	assert.Equal(expected, result)
}

func TestGetAssets(t *testing.T) {
	assert := assert.New(t)
	fs := afero.NewMemMapFs()

	type test struct {
		Directory      string
		Arguments      []string
		ExpectedAssets []asset.Asset
	}

	suite := map[string]test{
		"Functionality": {
			Directory: ".",
			Arguments: []string{
				"file1",
				"./file2 file3",
			},
			ExpectedAssets: []asset.Asset{
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
			},
		},
		"New Line Seperator": {
			Directory: ".",
			Arguments: []string{
				`file1
file2
file3
./file4`,
				`file5
file6`,
			},
			ExpectedAssets: []asset.Asset{
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
				{
					Name: "file5",
					Path: "file5",
				},
				{
					Name: "file6",
					Path: "file6",
				},
			},
		},
		"Comma Seperator": {
			Directory: ".",
			Arguments: []string{
				"file1,file2",
				"file3,./file4,file5",
				"./file6",
			},
			ExpectedAssets: []asset.Asset{
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
				{
					Name: "file5",
					Path: "file5",
				},
				{
					Name: "file6",
					Path: "file6",
				},
			},
		},
		"Pipe Seperator": {
			Directory: ".",
			Arguments: []string{
				"file1|file2",
				"file3|./file4|file5",
				"./file6",
			},
			ExpectedAssets: []asset.Asset{
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
				{
					Name: "file5",
					Path: "file5",
				},
				{
					Name: "file6",
					Path: "file6",
				},
			},
		},
		"Not Current Directory": {
			Directory: "workspace",
			Arguments: []string{
				"file1",
				"./file2",
			},
			ExpectedAssets: []asset.Asset{
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
		"Pattern Matching": {
			Directory: ".",
			Arguments: []string{
				"file*",
			},
			ExpectedAssets: []asset.Asset{
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
				{
					Name: "file5",
					Path: "file5",
				},
				{
					Name: "file6",
					Path: "file6",
				},
			},
		},
	}

	var counter int
	for name, test := range suite {
		counter++
		t.Logf("Test Case %v/%v - %s", counter, len(suite), name)

		for _, asset := range test.ExpectedAssets {
			_, err := fs.Create(asset.Path)
			assert.Equal(nil, err, fmt.Sprintf("preparation: error creating test file '%v'", asset.Path))
		}

		results := app.GetAssets(fs, test.Arguments)
		assert.Equal(test.ExpectedAssets, results)

		for _, asset := range test.ExpectedAssets {
			os.Remove(asset.Path)
		}
	}
}
