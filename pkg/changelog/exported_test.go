package changelog_test

import (
	"testing"

	"github.com/anton-yurchenko/git-release/pkg/changelog"
	"github.com/stretchr/testify/assert"
)

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
