package changelog_test

import (
	"fmt"
	"testing"

	"github.com/anton-yurchenko/git-release/pkg/changelog"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
)

func TestReadChanges(t *testing.T) {
	fs := afero.NewMemMapFs()
	file := createChangelog(fs, t)

	suite := map[string][]map[string]string{
		"pass": []map[string]string{
			map[string]string{
				"version":  "1.0.0",
				"expected": `- First stable release.`,
			},
			map[string]string{
				"version": "1.0.1-beta",
				"expected": `### Added
- New feature.

### Fixed
- Fixed env.var bug.`,
			},
		},
		"fail": []map[string]string{
			map[string]string{
				"version":  "99.0.0",
				"expected": ``,
			},
		},
	}

	for _, test := range suite["pass"] {
		assert := assert.New(t)

		m := changelog.Changes{
			File:    file,
			Version: test["version"],
			Body:    "",
		}

		err := m.ReadChanges(fs)

		assert.Equal(nil, err)
		assert.Equal(test["expected"], m.Body)
	}

	for _, test := range suite["fail"] {
		assert := assert.New(t)

		m := changelog.Changes{
			File:    file,
			Version: test["version"],
			Body:    "",
		}

		err := m.ReadChanges(fs)

		assert.EqualError(err, fmt.Sprintf("empty changelog for requested version: '%v'", test["version"]))
		assert.Equal(test["expected"], m.Body)
	}

	// in order to cover 100%, interface should have a private 'Read' method
	// i prefer to keep 'return err' uncovered.
	// TEST: err
	// assert := assert.New(t)
	// fs = afero.NewMemMapFs()
	// file = createChangelog(fs, t)

	// m := new(mocks.Changelog)

	// m.On("Read", fs).Return([]string{file}, errors.New("failure")).Once()

	// err := m.ReadChanges(fs)

	// assert.EqualError(err, "failure")
}

func TestSetFile(t *testing.T) {
	assert := assert.New(t)

	m := new(changelog.Changes)
	expected := "/home/user/file"
	m.SetFile(expected)

	assert.Equal(expected, m.File)
}

func TestGetFile(t *testing.T) {
	assert := assert.New(t)

	m := new(changelog.Changes)
	expected := "/home/user/file"
	m.SetFile(expected)

	assert.Equal(expected, m.GetFile())
}

func TestGetBody(t *testing.T) {
	assert := assert.New(t)

	expected := `### Added
- Feature A
- Feature B
- GitHub Actions as a CI system
- GitHub Release as an Artifactory system

### Changed
- User API

### Removed
- Previous CI build
- Previous Artifactory`

	m := changelog.Changes{
		Body: expected,
	}

	assert.Equal(expected, m.GetBody())
}
