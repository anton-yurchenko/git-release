package main

import (
	"bytes"
	"fmt"
	"os"
	"path"
	"reflect"
	"sort"
	"strings"
	"text/template"

	"git-release/env"
	"git-release/release"

	changelog "github.com/anton-yurchenko/go-changelog"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/afero"
)

// removed maps settings deleted in v7 to their replacement.
//
// These are fatal rather than ignored on purpose. The majority of users pin a
// floating reference (`:latest`, `@main`), so they cross a major boundary
// without editing their workflow; and a silently ignored RELEASE_NAME_PREFIX
// would publish a wrongly titled release under a green check. A fatal is also
// the only mechanism that reaches `uses: docker://...` users, for whom the
// runner performs no manifest processing at all - no defaults, no deprecation
// warnings, no unexpected-input errors.
var removed = map[string]string{
	"RELEASE_NAME":         "NAME_TEMPLATE, for example \"Ship It\"",
	"RELEASE_NAME_PREFIX":  "NAME_TEMPLATE, for example \"Release: {{ .Tag }}\"",
	"RELEASE_NAME_SUFFIX":  "NAME_TEMPLATE, for example \"{{ .Tag }} (nightly)\"",
	"RELEASE_NAME_POSTFIX": "NAME_TEMPLATE, for example \"{{ .Tag }} (nightly)\"",
	"ALLOW_TAG_PREFIX":     "TAG_PREFIX_REGEX",
}

// Configuration is a git-release settings struct
type Configuration struct {
	AllowEmptyChangelog bool
	UnreleasedCreate    bool
	UnreleasedDelete    bool
	TagPrefix           string
	ChangelogFile       string
	Assets              []string
	NameTemplate        *template.Template
	BodyTemplate        *template.Template
}

// templateData is the model exposed to NAME_TEMPLATE.
//
// NOTE: a struct and not a map, deliberately. text/template resolves struct
// fields through reflection and fails execution with `can't evaluate field X`
// for an unknown field, so a typo terminates the run. A map renders '<no value>'
// instead, and the `missingkey` option is a no-op on structs - it is consulted
// only for map receivers.
type templateData struct {
	Owner      string
	Repo       string
	Tag        string
	Version    string
	Major      string
	Minor      string
	Patch      string
	Prerelease string
	CommitHash string

	// Named Is* so that .Prerelease (the semver identifier, e.g. "rc.1") and
	// the boolean cannot be confused - they would otherwise differ by a single
	// letter's case in a case-sensitive lookup.
	IsDraft      bool
	IsPreRelease bool
	IsUnreleased bool
}

// bodyTemplateData extends templateData with everything only the body can see:
// the changelog, and the already-rendered name.
type bodyTemplateData struct {
	templateData
	Name      string
	Changelog string
}

// templateFields lists the fields a template may reference. Derived from the
// data model itself so the error message cannot drift away from the code.
func templateFields(v any) string {
	out := make([]string, 0)

	var walk func(reflect.Type)
	walk = func(t reflect.Type) {
		for i := 0; i < t.NumField(); i++ {
			if f := t.Field(i); f.Anonymous {
				walk(f.Type)
			} else {
				out = append(out, "."+f.Name)
			}
		}
	}
	walk(reflect.TypeOf(v))

	sort.Strings(out)

	return strings.Join(out, ", ")
}

func newTemplateData(rel *release.Release) templateData {
	d := templateData{
		IsDraft:      rel.Draft,
		IsPreRelease: rel.PreRelease,
	}

	if rel.Reference != nil {
		d.Tag = rel.Reference.Tag
		d.Version = rel.Reference.Version
		d.Major = rel.Reference.Major
		d.Minor = rel.Reference.Minor
		d.Patch = rel.Reference.Patch
		d.Prerelease = rel.Reference.Prerelease
		d.CommitHash = rel.Reference.CommitHash
		d.IsUnreleased = rel.Reference.Version == release.UnreleasedVersion
	}

	if rel.Slug != nil {
		d.Owner = rel.Slug.Owner
		d.Repo = rel.Slug.Name
	}

	return d
}

// parseTemplate compiles a user supplied template.
//
// Templates are parsed during configuration so that a malformed one fails
// before any file is read or any API call is made.
//
// No custom functions are registered, and none ever should be: a NAME_TEMPLATE
// is routinely built from untrusted runtime data (`${{ github.event.head_commit.message }}`),
// and with a scalar-only model and no functions that is inert. A `env` or
// `exec` function would turn the same primitive into a way to write GITHUB_TOKEN
// into a public release body. Adding a function later is backwards compatible;
// removing one is not.
func parseTemplate(name, text string) (*template.Template, error) {
	// {{.}} and {{$}} render the whole data structure, which becomes a leak the
	// day any sensitive field joins the model.
	for _, forbidden := range []string{"{{.}}", "{{ . }}", "{{$}}", "{{ $ }}"} {
		if strings.Contains(text, forbidden) {
			return nil, errors.Errorf("%v may not render the whole data structure, reference a field instead", name)
		}
	}

	t, err := template.New(name).Parse(text)
	if err != nil {
		return nil, errors.Wrapf(err, "error parsing %v", name)
	}

	return t, nil
}

// GetConfig sets validated Release/Changelog configuration
func GetConfig(fs afero.Fs) (*Configuration, error) {
	if err := rejectRemoved(); err != nil {
		return nil, err
	}

	conf := new(Configuration)

	var err error
	if conf.AllowEmptyChangelog, err = env.Bool("ALLOW_EMPTY_CHANGELOG"); err != nil {
		return nil, err
	}

	unreleased, err := env.Enum("UNRELEASED", "", "update", "delete")
	if err != nil {
		return nil, err
	}
	switch unreleased {
	case "update":
		conf.UnreleasedCreate = true
	case "delete":
		conf.UnreleasedDelete = true
	}

	conf.TagPrefix = env.Get("TAG_PREFIX_REGEX")
	conf.Assets = getAssets()

	if v := env.Get("NAME_TEMPLATE"); v != "" {
		if conf.NameTemplate, err = parseTemplate("NAME_TEMPLATE", v); err != nil {
			return nil, err
		}
	}

	if v := env.Get("BODY_TEMPLATE"); v != "" {
		if conf.BodyTemplate, err = parseTemplate("BODY_TEMPLATE", v); err != nil {
			return nil, err
		}
	}

	c := env.Get("CHANGELOG_FILE")
	if c == "" {
		c = "CHANGELOG.md"
	}
	conf.ChangelogFile = path.Join(os.Getenv("GITHUB_WORKSPACE"), c)

	b, err := afero.Exists(fs, conf.ChangelogFile)
	if err != nil {
		return nil, errors.Wrap(err, "error validating changelog file")
	}

	if !b {
		if c != "none" {
			log.Errorf("changelog file %v not found!", c)
		}

		conf.ChangelogFile = ""
	}

	return conf, nil
}

// getAssets returns the release assets, as configured by the action's `args`
// input - divided by a newline, space, comma or pipe, as documented since v1.
//
// NOTE: read from INPUT_ARGS rather than argv. The runner splices a container
// action's `args` onto the docker command line, where it is split on whitespace
// before the process starts, while the JavaScript wrapper passed the whole value
// as a single argument. The two distribution modes therefore disagreed about
// which assets to upload - the same input produced five assets through the
// container and two through the wrapper, with the difference silently dropped.
// INPUT_ARGS carries the value verbatim in both modes.
//
// All four separators are honoured together. v6 picked whichever matched first,
// so a newline-separated list containing a space was split on the space instead,
// which is the other half of the same bug.
func getAssets() []string {
	out := make([]string, 0)

	fields := strings.FieldsFunc(os.Getenv("INPUT_ARGS"), func(r rune) bool {
		switch r {
		case '\n', '\r', '\t', ' ', ',', '|':
			return true
		}

		return false
	})

	for _, f := range fields {
		if f != "" {
			out = append(out, f)
		}
	}

	return out
}

// rejectRemoved fails the run when a setting deleted in v7 is still configured.
func rejectRemoved() error {
	found := make([]string, 0)

	for name := range removed {
		if env.Get(name) != "" {
			found = append(found, name)
		}
	}

	if len(found) == 0 {
		return nil
	}

	sort.Strings(found)

	msg := new(strings.Builder)
	msg.WriteString("configuration removed in v7 is still in use:")
	for _, name := range found {
		fmt.Fprintf(msg, "\n- %v was replaced by %v", name, removed[name])
	}

	return errors.New(msg.String())
}

// GetName returns the release title.
func (c *Configuration) GetName(rel *release.Release) (string, error) {
	if c.NameTemplate == nil {
		return rel.Name, nil
	}

	var b bytes.Buffer
	if err := c.NameTemplate.Execute(&b, newTemplateData(rel)); err != nil {
		return "", errors.Wrapf(err, "error rendering NAME_TEMPLATE (supported fields: %v)", templateFields(templateData{}))
	}

	return b.String(), nil
}

// GetBody returns the release body: the changelog as-is, or the changelog
// rendered into BODY_TEMPLATE together with the release metadata.
func (c *Configuration) GetBody(fs afero.Fs, rel *release.Release) (string, error) {
	var body string

	if c.ChangelogFile != "" {
		var err error

		body, err = c.GetChangelog(fs, rel)
		if err != nil {
			return "", errors.Wrap(err, "error reading changelog")
		}
	}

	if c.BodyTemplate == nil {
		return body, nil
	}

	d := bodyTemplateData{
		templateData: newTemplateData(rel),
		Name:         rel.Name,
		Changelog:    body,
	}

	// NOTE: rendered into a buffer rather than straight to a destination, since
	// text/template writes incrementally and leaves partial output behind when
	// Execute fails.
	var b bytes.Buffer
	if err := c.BodyTemplate.Execute(&b, d); err != nil {
		return "", errors.Wrapf(err, "error rendering BODY_TEMPLATE (supported fields: %v)", templateFields(d))
	}

	return b.String(), nil
}

func (c *Configuration) GetChangelog(fs afero.Fs, rel *release.Release) (string, error) {
	p, err := changelog.NewParserWithFilesystem(fs, c.ChangelogFile)
	if err != nil {
		return "", errors.Wrap(err, "error loading changelog file")
	}

	changes, err := p.Parse()
	if err != nil {
		return "", errors.Wrap(err, "error parsing changelog file")
	}

	var msg string
	if rel.Reference.Version == release.UnreleasedVersion {
		if changes.Unreleased != nil && changes.Unreleased.Changes != nil {
			return changes.Unreleased.Changes.ToString(), nil
		} else {
			msg = "changelog file does not contain changes in Unreleased scope"
		}
	} else {
		r := changes.GetRelease(rel.Reference.Version)

		if r != nil {
			if r.Changes != nil {
				return r.Changes.ToString(), nil
			} else {
				msg = fmt.Sprintf("changelog file does not contain changes for version %v", rel.Reference.Version)
			}
		} else {
			msg = fmt.Sprintf("changelog file does not contain version %v", rel.Reference.Version)
		}
	}

	if !c.AllowEmptyChangelog {
		return "", errors.New(msg)
	}

	log.Warn(msg)
	return "", nil
}
