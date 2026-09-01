package env_test

import (
	"testing"

	"git-release/env"

	"github.com/stretchr/testify/assert"
)

func TestGet(t *testing.T) {
	a := assert.New(t)

	t.Run("unset", func(t *testing.T) {
		a.Equal("", env.Get("GIT_RELEASE_TEST_UNSET"))
	})

	t.Run("set", func(t *testing.T) {
		t.Setenv("NAME_TEMPLATE", "Release {{ .Tag }}")
		a.Equal("Release {{ .Tag }}", env.Get("NAME_TEMPLATE"))
	})

	// Settings are plain environment variables. The action input spelling is
	// deliberately NOT consulted: a bare `uses: docker://...` step never reads
	// action.yml, so supporting both would mean two spellings for one setting
	// with the benefits reaching only the JavaScript-wrapper minority.
	t.Run("input spelling is not consulted", func(t *testing.T) {
		t.Setenv("INPUT_NAME_TEMPLATE", "from-input")
		a.Equal("", env.Get("NAME_TEMPLATE"))
	})
}

func TestBool(t *testing.T) {
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
		"yes":         {Value: "yes", Error: true},
		"no":          {Value: "no", Error: true},
		"1":           {Value: "1", Error: true},
		"0":           {Value: "0", Error: true},
		"on":          {Value: "on", Error: true},
		"typo":        {Value: "ture", Error: true},
		"with spaces": {Value: " true ", Error: true},
	}

	for name, test := range suite {
		t.Run(name, func(t *testing.T) {
			t.Setenv("DRAFT_RELEASE", test.Value)

			got, err := env.Bool("DRAFT_RELEASE")
			if test.Error {
				a.Error(err)
				a.ErrorContains(err, "DRAFT_RELEASE")
				a.ErrorContains(err, test.Value)
				return
			}

			a.NoError(err)
			a.Equal(test.Expected, got)
		})
	}
}

func TestEnum(t *testing.T) {
	a := assert.New(t)

	t.Run("unset yields the fallback", func(t *testing.T) {
		v, err := env.Enum("UNRELEASED", "fallback", "update", "delete")
		a.NoError(err)
		a.Equal("fallback", v)
	})

	t.Run("exact match", func(t *testing.T) {
		t.Setenv("UNRELEASED", "delete")
		v, err := env.Enum("UNRELEASED", "", "update", "delete")
		a.NoError(err)
		a.Equal("delete", v)
	})

	// v6 compared UNRELEASED case-sensitively, so "Update" silently did nothing.
	t.Run("case insensitive", func(t *testing.T) {
		t.Setenv("UNRELEASED", "UpDaTe")
		v, err := env.Enum("UNRELEASED", "", "update", "delete")
		a.NoError(err)
		a.Equal("update", v)
	})

	t.Run("rejects an unknown value", func(t *testing.T) {
		t.Setenv("UNRELEASED", "remove")
		_, err := env.Enum("UNRELEASED", "", "update", "delete")
		a.ErrorContains(err, "UNRELEASED not supported")
		a.ErrorContains(err, "update, delete")
	})
}
