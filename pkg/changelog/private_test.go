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
	fs := afero.NewMemMapFs()

	// TEST: correct file
	t.Log("Test Case 1/2 - Functionality")
	file := createChangelog(fs, t)

	m := new(changelog.Changes)
	m.SetFile(file)

	result, err := m.Read(fs)
	expected := strings.Split(content, "\n")

	assert.Equal(nil, err)
	assert.Equal(expected, result)

	// TEST: read non existing file
	t.Log("Test Case 2/2 - File Does Not Exist")

	m = new(changelog.Changes)
	m.SetFile("./non-existing-file")

	result, err = m.Read(fs)

	assert.EqualError(err, "open non-existing-file: file does not exist")
	assert.Equal(make([]string, 0), result)
}

func TestGetEndOfFirstRelease(t *testing.T) {
	assert := assert.New(t)

	type test struct {
		Content  string
		Expected int
	}

	suite := map[string]test{
		"Functionality": {
			Content:  contentCases["functionality"],
			Expected: 46,
		},
		"Single Release": {
			Content:  contentCases["single-release"],
			Expected: 2,
		},
		"Empty": {
			Content:  contentCases["empty"],
			Expected: 1,
		},
		"Wrong Format": {
			Content:  contentCases["wrong-format"],
			Expected: 5,
		},
		"Inconsistent Format": {
			Content:  contentCases["inconsistent-format"],
			Expected: 8,
		},
		"Unreleased": {
			Content:  contentCases["unreleased"],
			Expected: 2,
		},
	}

	var counter int
	for name, test := range suite {
		counter++
		t.Logf("Test Case %v/%v - %s", counter, len(suite), name)

		result := changelog.GetEndOfFirstRelease(strings.Split(test.Content, "\n"))

		assert.Equal(test.Expected, result)
	}
}

func TestGetReleasesLines(t *testing.T) {
	assert := assert.New(t)

	type test struct {
		Content  string
		Expected []int
	}

	suite := map[string]test{
		"Functionality": {
			Content:  contentCases["functionality"],
			Expected: []int{4, 11, 14, 19, 34, 41, 44},
		},
		"Single Release": {
			Content:  contentCases["single-release"],
			Expected: []int{0},
		},
		"Empty": {
			Content:  contentCases["empty"],
			Expected: []int{},
		},
		"Wrong Format": {
			Content:  contentCases["wrong-format"],
			Expected: []int{},
		},
		"Inconsistent Format": {
			Content:  contentCases["inconsistent-format"],
			Expected: []int{3},
		},
		"Unreleased": {
			Content:  contentCases["unreleased"],
			Expected: []int{0},
		},
	}

	var counter int
	for name, test := range suite {
		counter++
		t.Logf("Test Case %v/%v - %s", counter, len(suite), name)

		result := changelog.GetReleasesLines(strings.Split(test.Content, "\n"))

		assert.Equal(test.Expected, result)
	}
}

func TestGetMargins(t *testing.T) {
	assert := assert.New(t)

	type test struct {
		Content  string
		Expected map[string]map[string]int
	}

	suite := map[string]test{
		"Functionality": {
			Content:  contentCases["functionality"],
			Expected: releasesContentMargins,
		},
		"Single Release": {
			Content: contentCases["single-release"],
			Expected: map[string]map[string]int{
				"1.0.0": {
					"start": 1,
					"end":   2,
				},
			},
		},
		"Empty": {
			Content: contentCases["empty"],
			Expected: map[string]map[string]int{
				"1.0.0": {},
			},
		},
		"Wrong Format": {
			Content: contentCases["wrong-format"],
			Expected: map[string]map[string]int{
				"1.0.0": {},
			},
		},
		"Inconsistent Format": {
			Content: contentCases["inconsistent-format"],
			Expected: map[string]map[string]int{
				"1.0.1": {
					"start": 4,
					"end":   8,
				},
				"1.0.0": {},
			},
		},
		"Unreleased": {
			Content: contentCases["unreleased"],
			Expected: map[string]map[string]int{
				"1.0.0": {
					"start": 1,
					"end":   2,
				},
				"0.0.9": {},
			},
		},
	}

	var counter int
	for name, test := range suite {
		counter++
		t.Logf("Test Case %v/%v - %s", counter, len(suite), name)

		for version, expected := range test.Expected {
			m := changelog.Changes{
				Version: version,
			}

			result := m.GetMargins(strings.Split(test.Content, "\n"))

			assert.Equal(expected, result)
		}
	}
}

func TestGetContent(t *testing.T) {
	assert := assert.New(t)

	type test struct {
		Content  string
		Margins  map[string]map[string]int
		Expected map[string]string
	}

	suite := map[string]test{
		"Functionality": {
			Content:  contentCases["functionality"],
			Margins:  releasesContentMargins,
			Expected: releasesContent,
		},
		"Single Release": {
			Content: contentCases["single-release"],
			Margins: map[string]map[string]int{
				"1.0.0": {
					"start": 1,
					"end":   2,
				},
			},
			Expected: map[string]string{
				"1.0.0": `- Release`,
			},
		},
		"Empty": {
			Content: contentCases["empty"],
			Margins: map[string]map[string]int{
				"1.0.0": {
					"start": 1,
					"end":   2,
				},
			},
			Expected: map[string]string{
				"1.0.0": ``,
			},
		},
		"Wrong Format": {
			Content: contentCases["wrong-format"],
			Margins: map[string]map[string]int{
				"1.0.1": {
					"start": 1,
					"end":   2,
				},
				"1.0.0": {
					"start": 4,
					"end":   5,
				},
			},
			Expected: map[string]string{
				"1.0.1": `- Fix`,
				"1.0.0": `- Release`,
			},
		},
		"Inconsistent Format": {
			Content: contentCases["inconsistent-format"],
			Margins: map[string]map[string]int{
				"1.0.2": {
					"start": 1,
					"end":   2,
				},
				"1.0.1": {
					"start": 4,
					"end":   5,
				},
				"1.0.0": {
					"start": 7,
					"end":   8,
				},
			},
			Expected: map[string]string{
				"1.0.2": `- Fix`,
				"1.0.1": `- Fix`,
				"1.0.0": `- Release`,
			},
		},
		"Unreleased": {
			Content: contentCases["unreleased"],
			Margins: map[string]map[string]int{
				"1.0.0": {
					"start": 1,
					"end":   2,
				},
				"0.0.9": {
					"start": 4,
					"end":   5,
				},
			},
			Expected: map[string]string{
				"1.0.0": `- Release`,
				"0.0.9": `[0.0.9]: https://github.com/...`,
			},
		},
	}

	var counter int
	for name, test := range suite {
		counter++
		t.Logf("Test Case %v/%v - %s", counter, len(suite), name)

		for version, margins := range test.Margins {

			expected := strings.Split(test.Expected[version], "\n")

			result := changelog.GetContent(margins, strings.Split(test.Content, "\n"))

			if test.Expected[version] == "" {
				assert.Equal(make([]string, 0), result)
			} else {
				assert.Equal(expected, result)
			}
		}
	}
}
