package app_test

import (
	"fmt"
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
	result, err := app.IsExists("./not-exist.zip", fs)

	assert.Equal(nil, err)
	assert.Equal(expected, result)

	// TEST: file exist
	expected = true

	file, err := fs.Create("./file1")
	file.Close()
	assert.Equal(nil, err, "preparation: error creating test file 'file1'")

	result, err = app.IsExists("./file1", fs)

	assert.Equal(nil, err)
	assert.Equal(expected, result)
}

func TestGetAssets(t *testing.T) {
	assert := assert.New(t)
	fs := afero.NewMemMapFs()

	baseDirs := []string{
		".",
		"workspace",
	}

	for _, baseDir := range baseDirs {
		for i := 0; i <= 6; i++ {
			_, err := fs.Create(fmt.Sprintf("%v/file%v", baseDir, i))
			assert.Equal(nil, err, fmt.Sprintf("preparation: error creating test file '%v/file%v'", baseDir, i))
		}
	}

	// TEST: arguments separated by space
	dir := "."

	args := []string{
		"file1",
		"./file2 file3",
	}

	expected := []asset.Asset{
		{
			Name: "file1",
			Path: "./file1",
		},
		{
			Name: "file2",
			Path: "./file2",
		},
		{
			Name: "file3",
			Path: "./file3",
		},
	}

	results := app.GetAssets(dir, fs, args)

	assert.Equal(expected, results)

	// TEST: arguments separated by new line
	dir = "."

	args = []string{
		`file1
file2
file3
./file4`,
		`file5
file6`,
	}

	expected = []asset.Asset{
		{
			Name: "file1",
			Path: "./file1",
		},
		{
			Name: "file2",
			Path: "./file2",
		},
		{
			Name: "file3",
			Path: "./file3",
		},
		{
			Name: "file4",
			Path: "./file4",
		},
		{
			Name: "file5",
			Path: "./file5",
		},
		{
			Name: "file6",
			Path: "./file6",
		},
	}

	results = app.GetAssets(dir, fs, args)

	assert.Equal(expected, results)

	// TEST: arguments separated by comma
	dir = "."

	args = []string{
		"file1,file2",
		"file3,./file4,file5",
		"./file6",
	}

	expected = []asset.Asset{
		{
			Name: "file1",
			Path: "./file1",
		},
		{
			Name: "file2",
			Path: "./file2",
		},
		{
			Name: "file3",
			Path: "./file3",
		},
		{
			Name: "file4",
			Path: "./file4",
		},
		{
			Name: "file5",
			Path: "./file5",
		},
		{
			Name: "file6",
			Path: "./file6",
		},
	}

	results = app.GetAssets(dir, fs, args)

	assert.Equal(expected, results)

	// TEST: arguments separated by pipe
	dir = "."

	args = []string{
		"file1|file2",
		"file3|./file4|file5",
		"./file6",
	}

	expected = []asset.Asset{
		{
			Name: "file1",
			Path: "./file1",
		},
		{
			Name: "file2",
			Path: "./file2",
		},
		{
			Name: "file3",
			Path: "./file3",
		},
		{
			Name: "file4",
			Path: "./file4",
		},
		{
			Name: "file5",
			Path: "./file5",
		},
		{
			Name: "file6",
			Path: "./file6",
		},
	}

	results = app.GetAssets(dir, fs, args)

	assert.Equal(expected, results)

	// TEST: arguments separated by space not in current directory
	dir = "workspace"

	args = []string{
		"file1",
		"./file2",
	}

	expected = []asset.Asset{
		{
			Name: "file1",
			Path: "workspace/file1",
		},
		{
			Name: "file2",
			Path: "workspace/file2",
		},
	}

	results = app.GetAssets(dir, fs, args)

	assert.Equal(expected, results)
}
