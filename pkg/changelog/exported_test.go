package changelog_test

import (
	"testing"

	"github.com/anton-yurchenko/git-release/pkg/changelog"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
)

func TestSetFile(t *testing.T) {
	assert := assert.New(t)
	t.Log("Test Case 1/1 - Functionality")

	m := new(changelog.Changes)
	expected := "/home/user/file"
	m.SetFile(expected)

	assert.Equal(expected, m.File)
}

func TestGetFile(t *testing.T) {
	assert := assert.New(t)
	t.Log("Test Case 1/1 - Functionality")

	m := new(changelog.Changes)
	expected := "/home/user/file"
	m.SetFile(expected)

	assert.Equal(expected, m.GetFile())
}

func TestGetBody(t *testing.T) {
	assert := assert.New(t)
	t.Log("Test Case 1/1 - Functionality")

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

func TestReadChanges(t *testing.T) {
	assert := assert.New(t)
	fs := afero.NewMemMapFs()

	content := `## [1.0.3] - 2014-08-09
### Added
- Feature

### Changed
- Behavior

## [1.0.2] - 2014-07-10
### Changed
- Behavior

## [1.0.1] - 2014-05-31
### Fixed
- Bug

[Unreleased]: https://github.com/anton-yurchenko/git-release/compare/v1.0.0...HEAD
[0.9.0]: https://github.com/anton-yurchenko/git-release/compare/v0.9.0...v0.8.3
[0.8.3]: https://github.com/anton-yurchenko/git-release/compare/v0.8.3...v0.8.2
`

	err := afero.WriteFile(fs, "CHANGELOG.md", []byte(content), 0644)
	assert.Equal(nil, err, "preparation: error creating test file 'CHANGELOG.md'")

	type test struct {
		Changes        changelog.Changes
		ExpectedError  string
		ExpectedResult string
	}

	suite := map[string]test{
		"Functionality 1": {
			Changes: changelog.Changes{
				File:    "CHANGELOG.md",
				Version: "1.0.3",
			},
			ExpectedError: "",
			ExpectedResult: `### Added
- Feature

### Changed
- Behavior`,
		},
		"Functionality 2": {
			Changes: changelog.Changes{
				File:    "CHANGELOG.md",
				Version: "1.0.2",
			},
			ExpectedError: "",
			ExpectedResult: `### Changed
- Behavior`,
		},
		"Functionality 3": {
			Changes: changelog.Changes{
				File:    "CHANGELOG.md",
				Version: "1.0.1",
			},
			ExpectedError: "",
			ExpectedResult: `### Fixed
- Bug`,
		},
		"Non Existing Version": {
			Changes: changelog.Changes{
				File:    "CHANGELOG.md",
				Version: "2.0.0",
			},
			ExpectedError:  "",
			ExpectedResult: "",
		},
		"Ignore Unreleased Versions": {
			Changes: changelog.Changes{
				File:    "CHANGELOG.md",
				Version: "0.9.0",
			},
			ExpectedError:  "",
			ExpectedResult: "",
		},
	}

	var counter int
	for name, test := range suite {
		counter++
		t.Logf("Test Case %v/%v - %s", counter, len(suite), name)

		err := test.Changes.ReadChanges(fs)

		if test.ExpectedError != "" {
			assert.EqualError(err, test.ExpectedError)
		} else {
			assert.Equal(nil, err)

			assert.Equal(test.ExpectedResult, test.Changes.Body)
		}
	}

}
