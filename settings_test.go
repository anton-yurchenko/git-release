package main

import (
	"io"
	"testing"

	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

// TestGetAssets pins the asset list contract.
//
// The list is read from INPUT_ARGS, which the runner sets from the action's
// `args` input in BOTH distribution modes. v6 read argv instead, which the
// runner had already whitespace-split for the container but not for the
// JavaScript wrapper, so the same input produced a different set of assets
// depending on how the action was referenced.
func TestGetAssets(t *testing.T) {
	a := assert.New(t)

	type test struct {
		Value    string
		Expected []string
	}

	suite := map[string]test{
		"unset":  {Value: "", Expected: []string{}},
		"single": {Value: "build/app.zip", Expected: []string{"build/app.zip"}},
		"newline separated": {
			Value:    "build/a.zip\nbuild/b.zip",
			Expected: []string{"build/a.zip", "build/b.zip"},
		},
		"space separated": {
			Value:    "build/a.zip build/b.zip",
			Expected: []string{"build/a.zip", "build/b.zip"},
		},
		"comma separated": {
			Value:    "build/a.zip,build/b.zip",
			Expected: []string{"build/a.zip", "build/b.zip"},
		},
		"pipe separated": {
			Value:    "build/a.zip|build/b.zip",
			Expected: []string{"build/a.zip", "build/b.zip"},
		},
		// v6 split on whichever separator matched first, so a comma-and-space
		// list lost everything after the first entry, and a newline list
		// containing a space was split on the space instead.
		"mixed separators": {
			Value:    "build/a.zip, build/b.zip\nbuild/c.zip|build/d.zip",
			Expected: []string{"build/a.zip", "build/b.zip", "build/c.zip", "build/d.zip"},
		},
		"blank lines and padding": {
			Value:    "\n  build/a.zip  \n\n build/b.zip \n",
			Expected: []string{"build/a.zip", "build/b.zip"},
		},
	}

	for name, test := range suite {
		t.Run(name, func(t *testing.T) {
			t.Setenv("INPUT_ARGS", test.Value)
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
	} {
		t.Run(name, func(t *testing.T) {
			t.Setenv(name, "anything")

			err := rejectRemoved()
			a.Error(err)
			a.ErrorContains(err, name)
			a.ErrorContains(err, replacement)
		})
	}

	t.Run("reports every offender at once", func(t *testing.T) {
		t.Setenv("RELEASE_NAME_PREFIX", "Release: ")
		t.Setenv("RELEASE_NAME_SUFFIX", " (nightly)")

		err := rejectRemoved()
		a.ErrorContains(err, "RELEASE_NAME_PREFIX")
		a.ErrorContains(err, "RELEASE_NAME_SUFFIX")
	})

	// `args` is NOT removed - it is still the action's input, only its delivery
	// changed from argv to INPUT_ARGS.
	t.Run("args is still accepted", func(t *testing.T) {
		t.Setenv("INPUT_ARGS", "build/*.zip")
		a.NoError(rejectRemoved())
	})
}
