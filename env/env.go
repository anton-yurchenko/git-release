// Package env reads git-release settings from the environment.
//
// Settings are environment variables, as they have been since v1. They are
// deliberately NOT declared as action inputs: `uses: docker://antonyurchenko/git-release:vN`
// is the documented primary usage, and a bare container reference never reads
// action.yml - no defaults, no `required`, no unexpected-input warnings. Every
// benefit of declaring inputs would reach only the JavaScript-wrapper minority,
// while the cost - two spellings for one setting, a precedence rule, and the
// support questions that follow - would land on everyone.
//
// The one exception is the asset list, which has always been the action's single
// input. See getAssets in config.go.
package env

import (
	"fmt"
	"os"
	"slices"
	"strings"
)

// Get returns the value of a setting.
func Get(name string) string {
	return os.Getenv(name)
}

// Bool returns a strictly parsed boolean setting.
//
// v6 treated any value other than "true" as false, so `DRAFT_RELEASE: "yes"`
// published a public release for someone who asked for a draft, silently.
// Anything that is neither true nor false is now an error.
func Bool(name string) (bool, error) {
	switch v := Get(name); strings.ToLower(v) {
	case "":
		return false, nil
	case "true":
		return true, nil
	case "false":
		return false, nil
	default:
		return false, fmt.Errorf("%v expects 'true' or 'false', received '%v'", name, v)
	}
}

// Enum returns a setting validated against the allowed values, case
// insensitively. An empty setting yields fallback.
func Enum(name, fallback string, allowed ...string) (string, error) {
	v := strings.ToLower(Get(name))
	if v == "" {
		return fallback, nil
	}

	if slices.Contains(allowed, v) {
		return v, nil
	}

	return "", fmt.Errorf("%v not supported, possible values are [%v]", name, strings.Join(allowed, ", "))
}
