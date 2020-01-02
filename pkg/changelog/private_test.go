package changelog_test

import (
	"strings"
	"testing"

	"github.com/anton-yurchenko/git-release/pkg/changelog"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
)

func TestRead(t *testing.T) {
	assert := assert.New(t)

	// TEST: correct file
	fs := afero.NewMemMapFs()
	file := createChangelog(fs, t)

	m := new(changelog.Changes)
	m.SetFile(file)

	result, err := m.Read(fs)
	expected := strings.Split(content, "\n")

	assert.Equal(nil, err)
	assert.Equal(expected, result)

	// TEST: read non existing file
	fs = afero.NewMemMapFs()

	m = new(changelog.Changes)
	m.SetFile("./non-existing-file")

	result, err = m.Read(fs)

	assert.EqualError(err, "open non-existing-file: file does not exist")
	assert.Equal(make([]string, 0), result)
}

func TestGetEndOfFirstRelease(t *testing.T) {
	assert := assert.New(t)

	// TEST: expected content
	expected := 46

	result := changelog.GetEndOfFirstRelease(strings.Split(content, "\n"))

	assert.Equal(expected, result)

	// TEST: single release changelog
	singleReleaseChangelog := `## [1.0.0] - 2018-01-01
- First stable release.`

	expected = 2

	result = changelog.GetEndOfFirstRelease(strings.Split(singleReleaseChangelog, "\n"))

	assert.Equal(expected, result)
}

func TestGetReleasesLines(t *testing.T) {
	assert := assert.New(t)

	expected := []int{4, 11, 14, 19, 34, 41, 44}
	result := changelog.GetReleasesLines(strings.Split(content, "\n"))

	assert.Equal(expected, result)
}

func TestGetMargins(t *testing.T) {
	for version, expected := range releasesContentMargins {
		assert := assert.New(t)

		m := changelog.Changes{
			Version: version,
		}

		result := m.GetMargins(strings.Split(content, "\n"))

		assert.Equal(expected, result)
	}
}

func TestGetContent(t *testing.T) {
	for version, margins := range releasesContentMargins {
		assert := assert.New(t)

		expected := strings.Split(releasesContent[version], "\n")
		result := changelog.GetContent(margins, strings.Split(content, "\n"))

		assert.Equal(expected, result)
	}
}
