package main

import (
	"os"
	"testing"

	"git-release/release"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
)

func init() {
	if err := os.Setenv("GITHUB_REPOSITORY", "owner/repo"); err != nil {
		panic(err)
	}
	if err := os.Setenv("GITHUB_TOKEN", "token"); err != nil {
		panic(err)
	}
	if err := os.Setenv("GITHUB_WORKSPACE", "."); err != nil {
		panic(err)
	}
	if err := os.Setenv("GITHUB_API_URL", "https://api.github.com"); err != nil {
		panic(err)
	}
	if err := os.Setenv("GITHUB_SERVER_URL", "https://github.com"); err != nil {
		panic(err)
	}
	if err := os.Setenv("GITHUB_REF", "refs/tags/1.0.0"); err != nil {
		panic(err)
	}
	if err := os.Setenv("GITHUB_SHA", "deadbeef"); err != nil {
		panic(err)
	}
}

func TestGetBody(t *testing.T) {
	a := assert.New(t)
	fs := afero.NewMemMapFs()

	// prepare changelog file
	changelog := `# Changelog

## [1.0.0] - 2024-01-01
### Added
- Something`
	a.NoError(afero.WriteFile(fs, "CHANGELOG.md", []byte(changelog), 0644))

	conf := &Configuration{ChangelogFile: "CHANGELOG.md", BodyTemplate: "Header\n{{.Changelog}}\nFooter"}
	rel := &release.Release{Reference: &release.Reference{Version: "1.0.0"}}

	body, err := conf.GetBody(fs, rel)
	a.NoError(err)
	expected := "Header\n### Added\n\n- Something\n\nFooter"
	a.Equal(expected, body)
}

func TestGetBodyMissingChangelog(t *testing.T) {
	a := assert.New(t)
	fs := afero.NewMemMapFs()
	conf := &Configuration{ChangelogFile: "", BodyTemplate: "Header\n{{.Changelog}}\nFooter"}
	rel := &release.Release{Reference: &release.Reference{Version: "1.0.0"}}

	body, err := conf.GetBody(fs, rel)
	a.NoError(err)
	a.Equal("Header\n\nFooter", body)
}
