package repository_test

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/anton-yurchenko/git-release/internal/pkg/repository"
	"github.com/stretchr/testify/assert"
)

func TestReadTag(t *testing.T) {
	assert := assert.New(t)

	type test struct {
		Ref           string
		Version       string
		AllowPrefix   bool
		ExpectedError string
	}

	suite := map[string]test{
		"Functionality": {
			Ref:           "refs/tags/v1.0.0",
			Version:       "1.0.0",
			AllowPrefix:   true,
			ExpectedError: "",
		},
		"Incorrect Environment Variable": {
			Ref:           "v1.0.0",
			Version:       "1.0.0",
			AllowPrefix:   true,
			ExpectedError: "malformed env.var 'GITHUB_REF' (control tag prefix via env.var 'ALLOW_TAG_PREFIX'): expected to match regex '^refs/tags/(?P<prefix>.*)(?P<major>0|[1-9]\\d*)\\.(?P<minor>0|[1-9]\\d*)\\.(?P<patch>0|[1-9]\\d*)(?:(?P<sep1>-)(?P<prerelease>(?:0|[1-9]\\d*|\\d*[a-zA-Z-][0-9a-zA-Z-]*)(?:\\.(?:0|[1-9]\\d*|\\d*[a-zA-Z-][0-9a-zA-Z-]*))*))?(?:(?P<sep2>\\+)(?P<buildmetadata>[0-9a-zA-Z-]+(?:\\.[0-9a-zA-Z-]+)*))?$', got 'v1.0.0'",
		},
		"Missing Environment Variable": {
			Ref:           "",
			Version:       "1.0.0",
			AllowPrefix:   true,
			ExpectedError: "env.var 'GITHUB_REF' is empty or not defined",
		},
		"Disabled AllowPrefix and Complex Version": {
			Ref:           "refs/tags/1.2.3----RC-SNAPSHOT.44.5.6--.77",
			Version:       "1.2.3----RC-SNAPSHOT.44.5.6--.77",
			AllowPrefix:   false,
			ExpectedError: "",
		},
	}

	var counter int
	for name, test := range suite {
		counter++
		t.Logf("Test Case %v/%v - %s", counter, len(suite), name)

		err := os.Setenv("GITHUB_REF", test.Ref)
		assert.Equal(nil, err, "preparation: error setting env.var 'GITHUB_REF'")

		var version string
		m := new(repository.Repository)
		err = m.ReadTag(&version, test.AllowPrefix)

		if test.ExpectedError != "" {
			assert.EqualError(err, test.ExpectedError)
		} else {
			assert.Equal(nil, err)

			assert.Equal(strings.TrimPrefix(test.Ref, "refs/tags/"), m.Tag)
			assert.Equal(test.Version, version)
		}

		err = os.Unsetenv("GITHUB_REF")
		assert.Equal(nil, err, "cleanup: error unsetting env.var 'GITHUB_REF'")
	}
}

func TestReadCommitHash(t *testing.T) {
	assert := assert.New(t)

	// TEST: env.var set
	t.Log("Test Case 1/2 - Functionality")
	expected := "123abc"

	err := os.Setenv("GITHUB_SHA", expected)
	assert.Equal(nil, err, "preparation: error setting env.var 'GITHUB_SHA'")

	m := new(repository.Repository)

	err = m.ReadCommitHash()

	assert.Equal(nil, err)
	assert.Equal(expected, m.CommitHash)

	// TEST: env.var not set
	t.Log("Test Case 1/2 - Missing Environment Variable")
	err = os.Setenv("GITHUB_SHA", "")
	assert.Equal(nil, err, "preparation: error setting env.var 'GITHUB_SHA'")

	m = new(repository.Repository)

	err = m.ReadCommitHash()

	assert.EqualError(err, "env.var 'GITHUB_SHA' is empty or not defined")
}

func TestReadProjectName(t *testing.T) {
	assert := assert.New(t)

	// TEST: env.var correct
	t.Log("Test Case 1/3 - Functionality")
	user := "user"
	project := "project"

	err := os.Setenv("GITHUB_REPOSITORY", fmt.Sprintf("%v/%v", user, project))
	assert.Equal(nil, err, "preparation: error setting env.var 'GITHUB_REPOSITORY'")

	m := new(repository.Repository)

	err = m.ReadProjectName()

	assert.Equal(nil, err)
	assert.Equal(user, m.Owner)
	assert.Equal(project, m.Project)

	// TEST: env.var incorrect
	t.Log("Test Case 2/3 - Incorrect Environmetal Variable")
	err = os.Setenv("GITHUB_REPOSITORY", "value")
	assert.Equal(nil, err, "preparation: error setting env.var 'GITHUB_REPOSITORY'")

	m = new(repository.Repository)

	err = m.ReadProjectName()

	assert.EqualError(err, "malformed env.var 'GITHUB_REPOSITORY': expected to match regex '^(?P<owner>[\\w,\\-,\\_\\.]+)\\/(?P<repo>[\\w\\,\\-\\_\\.]+)$', got 'value'")

	// TEST: env.var not set
	t.Log("Test Case 3/3 - Missing Environment Vriable")
	err = os.Setenv("GITHUB_REPOSITORY", "")
	assert.Equal(nil, err, "preparation: error setting env.var 'GITHUB_REPOSITORY'")

	m = new(repository.Repository)

	err = m.ReadProjectName()

	assert.EqualError(err, "env.var 'GITHUB_REPOSITORY' is empty or not defined")
}

func TestGetOwner(t *testing.T) {
	assert := assert.New(t)
	t.Log("Test Case 1/1 - Functionality")

	expected := "user"

	m := repository.Repository{
		Owner: expected,
	}

	result := m.GetOwner()

	assert.Equal(expected, result)
}

func TestGetProject(t *testing.T) {
	assert := assert.New(t)
	t.Log("Test Case 1/1 - Functionality")

	expected := "project"

	m := repository.Repository{
		Project: expected,
	}

	result := m.GetProject()

	assert.Equal(expected, result)
}

func TestGetTag(t *testing.T) {
	assert := assert.New(t)
	t.Log("Test Case 1/1 - Functionality")

	expected := "1.0.0"

	m := repository.Repository{
		Tag: expected,
	}

	result := m.GetTag()

	assert.Equal(expected, *result)
}

func TestGetCommitHash(t *testing.T) {
	assert := assert.New(t)
	t.Log("Test Case 1/1 - Functionality")

	expected := "123"

	m := repository.Repository{
		CommitHash: expected,
	}

	result := m.GetCommitHash()

	assert.Equal(expected, *result)
}
