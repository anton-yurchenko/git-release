package main

import (
	"io"
	"testing"

	"git-release/env"

	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

// TestSettingSpellings pins the dual-read contract: `with: draft_release:` and
// `env: DRAFT_RELEASE:` are the same setting, and the input spelling wins.
func TestSettingSpellings(t *testing.T) {
	a := assert.New(t)

	t.Run("env only", func(t *testing.T) {
		t.Setenv("NAME_TEMPLATE", "from-env")
		a.Equal("from-env", env.Get("NAME_TEMPLATE"))
	})

	t.Run("input only", func(t *testing.T) {
		t.Setenv("INPUT_NAME_TEMPLATE", "from-input")
		a.Equal("from-input", env.Get("NAME_TEMPLATE"))
	})

	t.Run("input wins", func(t *testing.T) {
		t.Setenv("NAME_TEMPLATE", "from-env")
		t.Setenv("INPUT_NAME_TEMPLATE", "from-input")
		a.Equal("from-input", env.Get("NAME_TEMPLATE"))
	})

	// The runner injects an empty INPUT_<name> for every declared input that has
	// no default, so on the wrapper path INPUT_<name> is always present. An
	// empty one must not shadow the environment fallback.
	t.Run("empty input does not shadow env", func(t *testing.T) {
		t.Setenv("NAME_TEMPLATE", "from-env")
		t.Setenv("INPUT_NAME_TEMPLATE", "")
		a.Equal("from-env", env.Get("NAME_TEMPLATE"))
	})
}

func TestBooleanSetting(t *testing.T) {
	a := assert.New(t)

	type test struct {
		Value    string
		Expected bool
		Error    bool
	}

	suite := map[string]test{
		"unset": {Value: "", Expected: false},
		"true":  {Value: "true", Expected: true},
		"True":  {Value: "True", Expected: true},
		"TRUE":  {Value: "TRUE", Expected: true},
		"false": {Value: "false", Expected: false},
		"False": {Value: "False", Expected: false},
		// v6 treated every one of these as false, so DRAFT_RELEASE: "yes"
		// published a public release for a user who asked for a draft.
		"yes":  {Value: "yes", Error: true},
		"1":    {Value: "1", Error: true},
		"on":   {Value: "on", Error: true},
		"typo": {Value: "ture", Error: true},
	}

	for name, test := range suite {
		t.Run(name, func(t *testing.T) {
			t.Setenv("DRAFT_RELEASE", test.Value)

			got, err := env.Bool("DRAFT_RELEASE")
			if test.Error {
				a.Error(err)
				a.ErrorContains(err, "DRAFT_RELEASE")
				return
			}

			a.NoError(err)
			a.Equal(test.Expected, got)
		})
	}
}

func TestEnumSetting(t *testing.T) {
	a := assert.New(t)

	t.Run("fallback when unset", func(t *testing.T) {
		v, err := env.Enum("PRE_RELEASE", "auto", "auto", "true", "false")
		a.NoError(err)
		a.Equal("auto", v)
	})

	t.Run("case insensitive", func(t *testing.T) {
		t.Setenv("UNRELEASED", "Update")
		v, err := env.Enum("UNRELEASED", "", "update", "delete")
		a.NoError(err)
		a.Equal("update", v)
	})

	t.Run("rejects unknown", func(t *testing.T) {
		t.Setenv("UNRELEASED", "remove")
		_, err := env.Enum("UNRELEASED", "", "update", "delete")
		a.ErrorContains(err, "UNRELEASED not supported")
	})
}

func TestGetAssets(t *testing.T) {
	a := assert.New(t)

	type test struct {
		Value    string
		Expected []string
	}

	suite := map[string]test{
		"unset":           {Value: "", Expected: []string{}},
		"single":          {Value: "build/app.zip", Expected: []string{"build/app.zip"}},
		"newline list":    {Value: "build/a.zip\nbuild/b.zip", Expected: []string{"build/a.zip", "build/b.zip"}},
		"blank lines":     {Value: "\nbuild/a.zip\n\n  \nbuild/b.zip\n", Expected: []string{"build/a.zip", "build/b.zip"}},
		"trailing spaces": {Value: "  build/a.zip  \n build/b.zip", Expected: []string{"build/a.zip", "build/b.zip"}},
		// v6 split on spaces, so a path containing one could never be expressed.
		"path with space": {Value: "build/release notes.txt", Expected: []string{"build/release notes.txt"}},
		// v6 split on commas and pipes too, silently mangling these names.
		"path with comma": {Value: "build/a,b.zip", Expected: []string{"build/a,b.zip"}},
		"path with pipe":  {Value: "build/a|b.zip", Expected: []string{"build/a|b.zip"}},
	}

	for name, test := range suite {
		t.Run(name, func(t *testing.T) {
			t.Setenv("ASSETS", test.Value)
			a.Equal(test.Expected, getAssets())
		})
	}
}

func TestRejectRemoved(t *testing.T) {
	a := assert.New(t)
	log.SetOutput(io.Discard)

	t.Run("clean environment passes", func(t *testing.T) {
		a.NoError(rejectRemoved())
	})

	for name, replacement := range map[string]string{
		"RELEASE_NAME":         "NAME_TEMPLATE",
		"RELEASE_NAME_PREFIX":  "NAME_TEMPLATE",
		"RELEASE_NAME_SUFFIX":  "NAME_TEMPLATE",
		"RELEASE_NAME_POSTFIX": "NAME_TEMPLATE",
		"ALLOW_TAG_PREFIX":     "TAG_PREFIX_REGEX",
		"ARGS":                 "ASSETS",
	} {
		t.Run(name, func(t *testing.T) {
			t.Setenv(name, "anything")

			err := rejectRemoved()
			a.Error(err)
			a.ErrorContains(err, name)
			a.ErrorContains(err, replacement)
		})
	}

	// A removed setting supplied as an action input must be caught too - that is
	// the spelling a `uses: docker://...` user would reach for, and the runner
	// performs no input validation at all for that reference form.
	t.Run("input spelling", func(t *testing.T) {
		t.Setenv("INPUT_RELEASE_NAME_PREFIX", "Release: ")

		err := rejectRemoved()
		a.ErrorContains(err, "RELEASE_NAME_PREFIX")
	})

	t.Run("reports every offender at once", func(t *testing.T) {
		t.Setenv("RELEASE_NAME_PREFIX", "Release: ")
		t.Setenv("RELEASE_NAME_SUFFIX", " (nightly)")

		err := rejectRemoved()
		a.ErrorContains(err, "RELEASE_NAME_PREFIX")
		a.ErrorContains(err, "RELEASE_NAME_SUFFIX")
	})
}
