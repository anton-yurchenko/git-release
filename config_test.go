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
