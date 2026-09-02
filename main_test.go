package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestValidateEnvironment pins the variables the action cannot run without.
// All seven are supplied by the Actions runner; a missing one means the action
// is being invoked outside a workflow, or the token was not passed through.
func TestValidateEnvironment(t *testing.T) {
	a := assert.New(t)

	required := []string{
		"GITHUB_REPOSITORY",
		"GITHUB_TOKEN",
		"GITHUB_WORKSPACE",
		"GITHUB_API_URL",
		"GITHUB_SERVER_URL",
		"GITHUB_REF",
		"GITHUB_SHA",
	}

	set := func(t *testing.T) {
		t.Helper()
		for _, v := range required {
			t.Setenv(v, "value")
		}
	}

	t.Run("all present", func(t *testing.T) {
		set(t)
		a.NoError(validateEnvironment())
	})

	for _, missing := range required {
		t.Run("missing "+missing, func(t *testing.T) {
			set(t)
			t.Setenv(missing, "")

			a.EqualError(validateEnvironment(), missing+" is not defined")
		})
	}
}
