// Package env reads git-release settings from the environment.
//
// Every setting has two spellings that mean the same thing: an action input
// (`with: draft_release:`) and a bare environment variable (`env: DRAFT_RELEASE:`).
// The rule is one sentence - the environment name is the ASCII-uppercase of the
// input name - which is exactly the transform the Actions runner applies when
// it exports `with:` keys as INPUT_*.
//
// Both spellings are supported because neither one works everywhere:
//
//   - `uses: docker://antonyurchenko/git-release:vN`, the documented primary
//     usage, is a bare container-registry reference. The runner never reads
//     action.yml for it, so declared inputs get no defaults, no `required`
//     enforcement and no unexpected-input warnings - but `with:` keys ARE
//     exported into the container as INPUT_*.
//   - `uses: anton-yurchenko/git-release@vN` runs the JavaScript wrapper, gets
//     full manifest processing, and is where `with:` is idiomatic.
//
// Reading both keeps the two distribution modes from diverging.
package env

import (
	"fmt"
	"os"
	"strings"
)

// Get returns the value of a setting. The action-input spelling wins.
//
// NOTE: an empty INPUT_<name> counts as unset rather than as an empty value.
// The runner injects an empty INPUT_<name> for every declared input that has no
// default, so on the wrapper path INPUT_<name> is ALWAYS present - testing
// presence instead of emptiness would silently disable the env fallback for
// every wrapper user.
func Get(name string) string {
	if v := os.Getenv("INPUT_" + name); v != "" {
		return v
	}

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

	for _, a := range allowed {
		if v == a {
			return v, nil
		}
	}

	return "", fmt.Errorf("%v not supported, possible values are [%v]", name, strings.Join(allowed, ", "))
}
