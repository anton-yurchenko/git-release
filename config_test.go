package main

import (
	"io"
	"testing"
	"text/template"

	"git-release/release"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
)

const changelogFile string = "CHANGELOG.md"

const changelogFixture string = `# Changelog

## [Unreleased]

## [1.0.0] - 2024-01-01

### Added

- Something
`

// changelogFixtureWithActions contains a Go template action inside a changelog
// entry, to prove a changelog cannot inject template actions into the body.
const changelogFixtureWithActions string = `# Changelog

## [1.0.0] - 2024-01-01

### Added

- Released {{.Tag}}
`

func testRelease() *release.Release {
	return &release.Release{
		Name: "v1.0.0",
		Slug: &release.Slug{Owner: "anton-yurchenko", Name: "git-release"},
		Reference: &release.Reference{
			CommitHash: "deadbeef",
			Tag:        "v1.0.0",
			Version:    "1.0.0",
			Major:      "1",
			Minor:      "0",
			Patch:      "0",
		},
		PreRelease: true,
	}
}

func testFilesystem(t *testing.T, content string) afero.Fs {
	t.Helper()

	fs := afero.NewMemMapFs()
	if err := afero.WriteFile(fs, changelogFile, []byte(content), 0o644); err != nil {
		t.Fatalf("error preparing test case: %v", err)
	}

	return fs
}

func mustTemplate(t *testing.T, name, text string) *template.Template {
	t.Helper()

	tmpl, err := parseTemplate(name, text)
	if err != nil {
		t.Fatalf("error preparing test case: %v", err)
	}

	return tmpl
}

func TestGetBody(t *testing.T) {
	a := assert.New(t)
	log.SetOutput(io.Discard)

	fs := testFilesystem(t, changelogFixture)

	// NOTE: the expected changelog is computed through the production code path
	// on purpose, so the assertions are not pinned to go-changelog's whitespace.
	changelog, err := (&Configuration{ChangelogFile: changelogFile}).GetChangelog(fs, testRelease())
	a.NoError(err)
	a.NotEmpty(changelog)

	type test struct {
		ChangelogFile       string
		AllowEmptyChangelog bool
		BodyTemplate        string
		Version             string
		Expected            string
		ExpectedError       []string
	}

	suite := map[string]test{
		"No Template - Changelog Passthrough": {
			ChangelogFile: changelogFile,
			Expected:      changelog,
		},
		"No Template And No Changelog File": {
			Expected: "",
		},
		"Changelog Injected Into Template": {
			ChangelogFile: changelogFile,
			BodyTemplate:  "Header\n{{.Changelog}}\nFooter",
			Expected:      "Header\n" + changelog + "\nFooter",
		},
		"All Fields Rendered": {
			BodyTemplate: "{{.Version}}|{{.Tag}}|{{.Name}}|{{.Owner}}|{{.Repo}}|{{.CommitHash}}|{{.Major}}.{{.Minor}}.{{.Patch}}|{{.IsDraft}}|{{.IsPreRelease}}",
			Expected:     "1.0.0|v1.0.0|v1.0.0|anton-yurchenko|git-release|deadbeef|1.0.0|false|true",
		},
		"Booleans Usable In Conditionals": {
			BodyTemplate: "{{if .IsPreRelease}}pre{{end}}{{if .IsDraft}}draft{{end}}",
			Expected:     "pre",
		},
		"Template Without Changelog File": {
			BodyTemplate: "Release {{.Tag}}",
			Expected:     "Release v1.0.0",
		},
		"Allow Empty Changelog With Template": {
			ChangelogFile:       changelogFile,
			AllowEmptyChangelog: true,
			BodyTemplate:        "Header\n{{.Changelog}}\nFooter",
			Version:             "9.9.9",
			Expected:            "Header\n\nFooter",
		},
		"Unknown Field Fails The Run": {
			BodyTemplate:  "A{{.Chnagelog}}B",
			ExpectedError: []string{"error rendering BODY_TEMPLATE", "can't evaluate field Chnagelog", ".Changelog"},
		},
		"Changelog Error Propagates": {
			ChangelogFile: changelogFile,
			BodyTemplate:  "{{.Changelog}}",
			Version:       "9.9.9",
			ExpectedError: []string{"error reading changelog: changelog file does not contain version 9.9.9"},
		},
	}

	var counter int
	for name, test := range suite {
		counter++
		t.Logf("Test Case %v/%v - %s", counter, len(suite), name)

		conf := &Configuration{
			ChangelogFile:       test.ChangelogFile,
			AllowEmptyChangelog: test.AllowEmptyChangelog,
		}
		if test.BodyTemplate != "" {
			conf.BodyTemplate = mustTemplate(t, "BODY_TEMPLATE", test.BodyTemplate)
		}

		rel := testRelease()
		if test.Version != "" {
			rel.Reference.Version = test.Version
		}

		body, err := conf.GetBody(fs, rel)

		if len(test.ExpectedError) > 0 {
			a.Error(err)

			for _, m := range test.ExpectedError {
				a.ErrorContains(err, m)
			}

			// NOTE: text/template writes incrementally, an error must not leak
			// a partial body.
			a.Equal("", body)

			continue
		}

		a.NoError(err)
		a.Equal(test.Expected, body)
	}
}

func TestGetName(t *testing.T) {
	a := assert.New(t)

	type test struct {
		NameTemplate  string
		Unreleased    bool
		Expected      string
		ExpectedError []string
	}

	suite := map[string]test{
		"No Template Keeps The Default": {
			Expected: "v1.0.0",
		},
		"Prefix And Suffix Compose Around The Tag": {
			NameTemplate: "Release: {{ .Tag }} (stable)",
			Expected:     "Release: v1.0.0 (stable)",
		},
		"Explicit Name": {
			NameTemplate: "Codename Ragnarok",
			Expected:     "Codename Ragnarok",
		},
		"Semantic Version Components": {
			NameTemplate: "{{ .Major }}.{{ .Minor }} series",
			Expected:     "1.0 series",
		},
		"Unreleased Branch": {
			NameTemplate: "{{ if .IsUnreleased }}Nightly{{ else }}{{ .Tag }}{{ end }}",
			Unreleased:   true,
			Expected:     "Nightly",
		},
		"Unknown Field Fails The Run": {
			NameTemplate:  "{{ .Chnagelog }}",
			ExpectedError: []string{"error rendering NAME_TEMPLATE", "can't evaluate field Chnagelog"},
		},
		// The name is rendered before the changelog is read, so .Changelog is
		// deliberately absent from its data model.
		"Changelog Is Not Available To The Name": {
			NameTemplate:  "{{ .Changelog }}",
			ExpectedError: []string{"can't evaluate field Changelog"},
		},
	}

	var counter int
	for name, test := range suite {
		counter++
		t.Logf("Test Case %v/%v - %s", counter, len(suite), name)

		conf := new(Configuration)
		if test.NameTemplate != "" {
			conf.NameTemplate = mustTemplate(t, "NAME_TEMPLATE", test.NameTemplate)
		}

		rel := testRelease()
		if test.Unreleased {
			rel.Reference.Version = release.UnreleasedVersion
			rel.Name = "Latest"
		}

		got, err := conf.GetName(rel)

		if len(test.ExpectedError) > 0 {
			a.Error(err)
			for _, m := range test.ExpectedError {
				a.ErrorContains(err, m)
			}
			continue
		}

		a.NoError(err)
		a.Equal(test.Expected, got)
	}
}

func TestGetBodyChangelogIsNotTemplated(t *testing.T) {
	a := assert.New(t)
	log.SetOutput(io.Discard)

	fs := testFilesystem(t, changelogFixtureWithActions)
	conf := &Configuration{
		ChangelogFile: changelogFile,
		BodyTemplate:  mustTemplate(t, "BODY_TEMPLATE", "{{.Changelog}}"),
	}

	body, err := conf.GetBody(fs, testRelease())
	a.NoError(err)
	a.Contains(body, "{{.Tag}}")
	a.NotContains(body, "v1.0.0")
}

func TestGetBodyNilReferenceAndSlug(t *testing.T) {
	a := assert.New(t)
	log.SetOutput(io.Discard)

	conf := &Configuration{
		BodyTemplate: mustTemplate(t, "BODY_TEMPLATE", "{{.Owner}}/{{.Repo}}@{{.Tag}}"),
	}

	body, err := conf.GetBody(afero.NewMemMapFs(), &release.Release{})
	a.NoError(err)
	a.Equal("/@", body)
}

func TestParseTemplate(t *testing.T) {
	a := assert.New(t)

	type test struct {
		Template      string
		ExpectedError string
	}

	suite := map[string]test{
		"Valid":                     {Template: "{{ .Tag }}", ExpectedError: ""},
		"Unclosed Action":           {Template: "{{ .Tag", ExpectedError: "unclosed action"},
		"Unknown Function":          {Template: "{{ upper .Tag }}", ExpectedError: `function "upper" not defined`},
		"Whole Structure":           {Template: "{{.}}", ExpectedError: "may not render the whole data structure"},
		"Whole Structure Spaced":    {Template: "{{ . }}", ExpectedError: "may not render the whole data structure"},
		"Whole Structure As Dollar": {Template: "{{$}}", ExpectedError: "may not render the whole data structure"},
	}

	var counter int
	for name, test := range suite {
		counter++
		t.Logf("Test Case %v/%v - %s", counter, len(suite), name)

		_, err := parseTemplate("NAME_TEMPLATE", test.Template)
		if test.ExpectedError == "" {
			a.NoError(err)
			continue
		}

		a.ErrorContains(err, test.ExpectedError)
	}
}

func TestTemplateFields(t *testing.T) {
	a := assert.New(t)

	// The body model must expose everything the name model does, plus its own
	// two fields - embedding is what guarantees exactly one spelling per field.
	name := templateFields(templateData{})
	body := templateFields(bodyTemplateData{})

	a.Contains(name, ".Tag")
	a.Contains(name, ".IsPreRelease")
	a.NotContains(name, ".Changelog")

	a.Contains(body, ".Changelog")
	a.Contains(body, ".Name")
	a.Contains(body, ".Tag")
}

func TestGetConfig(t *testing.T) {
	a := assert.New(t)
	log.SetOutput(io.Discard)

	type test struct {
		Env           map[string]string
		Files         map[string]string
		Expected      Configuration
		HasName       bool
		HasBody       bool
		ExpectedError string
	}

	suite := map[string]test{
		"Defaults": {
			Files:    map[string]string{changelogFile: changelogFixture},
			Expected: Configuration{ChangelogFile: changelogFile},
		},
		"Missing changelog file is tolerated": {
			Expected: Configuration{ChangelogFile: ""},
		},
		// 'none' is not a filename - it silences the "not found" error and
		// leaves the release body empty.
		"CHANGELOG_FILE none": {
			Env:      map[string]string{"CHANGELOG_FILE": "none"},
			Expected: Configuration{ChangelogFile: ""},
		},
		"CHANGELOG_FILE custom path": {
			Env:      map[string]string{"CHANGELOG_FILE": "docs/RELEASES.md"},
			Files:    map[string]string{"docs/RELEASES.md": changelogFixture},
			Expected: Configuration{ChangelogFile: "docs/RELEASES.md"},
		},
		"ALLOW_EMPTY_CHANGELOG": {
			Env:      map[string]string{"ALLOW_EMPTY_CHANGELOG": "true"},
			Files:    map[string]string{changelogFile: changelogFixture},
			Expected: Configuration{AllowEmptyChangelog: true, ChangelogFile: changelogFile},
		},
		"ALLOW_EMPTY_CHANGELOG rejects a non boolean": {
			Env:           map[string]string{"ALLOW_EMPTY_CHANGELOG": "yes"},
			ExpectedError: "ALLOW_EMPTY_CHANGELOG expects 'true' or 'false', received 'yes'",
		},
		"UNRELEASED update": {
			Files:    map[string]string{changelogFile: changelogFixture},
			Env:      map[string]string{"UNRELEASED": "update"},
			Expected: Configuration{UnreleasedCreate: true, ChangelogFile: changelogFile},
		},
		"UNRELEASED delete": {
			Files:    map[string]string{changelogFile: changelogFixture},
			Env:      map[string]string{"UNRELEASED": "delete"},
			Expected: Configuration{UnreleasedDelete: true, ChangelogFile: changelogFile},
		},
		"UNRELEASED rejects an unknown mode": {
			Env:           map[string]string{"UNRELEASED": "remove"},
			ExpectedError: "UNRELEASED not supported, possible values are [update, delete]",
		},
		"TAG_PREFIX_REGEX": {
			Files:    map[string]string{changelogFile: changelogFixture},
			Env:      map[string]string{"TAG_PREFIX_REGEX": "[a-z-]*"},
			Expected: Configuration{TagPrefix: "[a-z-]*", ChangelogFile: changelogFile},
		},
		"Assets": {
			Files:    map[string]string{changelogFile: changelogFixture},
			Env:      map[string]string{"INPUT_ARGS": "build/a.zip\nbuild/b.zip"},
			Expected: Configuration{ChangelogFile: changelogFile, Assets: []string{"build/a.zip", "build/b.zip"}},
		},
		"Templates are parsed": {
			Files: map[string]string{changelogFile: changelogFixture},
			Env: map[string]string{
				"NAME_TEMPLATE": "Release {{ .Tag }}",
				"BODY_TEMPLATE": "{{ .Changelog }}",
			},
			Expected: Configuration{ChangelogFile: changelogFile},
			HasName:  true,
			HasBody:  true,
		},
		// Templates are compiled during configuration so a typo fails before any
		// file is read or any API call is made.
		"Malformed NAME_TEMPLATE": {
			Env:           map[string]string{"NAME_TEMPLATE": "{{ .Tag"},
			ExpectedError: "error parsing NAME_TEMPLATE",
		},
		"Malformed BODY_TEMPLATE": {
			Env:           map[string]string{"BODY_TEMPLATE": "{{ upper .Tag }}"},
			ExpectedError: "error parsing BODY_TEMPLATE",
		},
		"A setting removed in v7 is fatal": {
			Env:           map[string]string{"RELEASE_NAME_PREFIX": "Release: "},
			ExpectedError: "RELEASE_NAME_PREFIX was replaced by NAME_TEMPLATE",
		},
	}

	// NOTE: subtests, not a flat loop - t.Setenv restores at the end of the TEST,
	// so a flat loop would leak each case's settings into every later case.
	for name, test := range suite {
		t.Run(name, func(t *testing.T) {
			for k, v := range test.Env {
				t.Setenv(k, v)
			}

			fs := afero.NewMemMapFs()
			for path, content := range test.Files {
				if err := afero.WriteFile(fs, path, []byte(content), 0o644); err != nil {
					t.Fatalf("error preparing test case: %v", err)
				}
			}

			conf, err := GetConfig(fs)

			if test.ExpectedError != "" {
				a.Error(err)
				a.ErrorContains(err, test.ExpectedError)
				return
			}

			if !a.NoError(err) {
				return
			}

			a.Equal(test.Expected.ChangelogFile, conf.ChangelogFile)
			a.Equal(test.Expected.AllowEmptyChangelog, conf.AllowEmptyChangelog)
			a.Equal(test.Expected.UnreleasedCreate, conf.UnreleasedCreate)
			a.Equal(test.Expected.UnreleasedDelete, conf.UnreleasedDelete)
			a.Equal(test.Expected.TagPrefix, conf.TagPrefix)

			if test.Expected.Assets != nil {
				a.Equal(test.Expected.Assets, conf.Assets)
			}

			a.Equal(test.HasName, conf.NameTemplate != nil)
			a.Equal(test.HasBody, conf.BodyTemplate != nil)
		})
	}
}

func TestGetChangelog(t *testing.T) {
	a := assert.New(t)
	log.SetOutput(io.Discard)

	const withUnreleased = `# Changelog

## [Unreleased]

### Added

- Something new

## [1.0.0] - 2024-01-01

### Added

- Something
`

	type test struct {
		Fixture             string
		Version             string
		AllowEmptyChangelog bool
		Expected            string
		ExpectedError       string
	}

	suite := map[string]test{
		"Released version": {
			Fixture:  changelogFixture,
			Version:  "1.0.0",
			Expected: "### Added\n\n- Something\n",
		},
		"Unreleased scope": {
			Fixture:  withUnreleased,
			Version:  release.UnreleasedVersion,
			Expected: "### Added\n\n- Something new\n",
		},
		"Unreleased scope with no entries": {
			Fixture:       changelogFixture,
			Version:       release.UnreleasedVersion,
			ExpectedError: "changelog file does not contain changes in Unreleased scope",
		},
		"Unreleased scope with no entries is tolerated": {
			Fixture:             changelogFixture,
			Version:             release.UnreleasedVersion,
			AllowEmptyChangelog: true,
			Expected:            "",
		},
		"Version absent from the changelog": {
			Fixture:       changelogFixture,
			Version:       "9.9.9",
			ExpectedError: "changelog file does not contain version 9.9.9",
		},
		"Version absent is tolerated": {
			Fixture:             changelogFixture,
			Version:             "9.9.9",
			AllowEmptyChangelog: true,
			Expected:            "",
		},
	}

	var counter int
	for name, test := range suite {
		counter++
		t.Logf("Test Case %v/%v - %s", counter, len(suite), name)

		conf := &Configuration{
			ChangelogFile:       changelogFile,
			AllowEmptyChangelog: test.AllowEmptyChangelog,
		}

		rel := testRelease()
		rel.Reference.Version = test.Version

		got, err := conf.GetChangelog(testFilesystem(t, test.Fixture), rel)

		if test.ExpectedError != "" {
			a.EqualError(err, test.ExpectedError)
			continue
		}

		a.NoError(err)
		a.Equal(test.Expected, got)
	}
}
