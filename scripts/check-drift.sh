#!/usr/bin/env bash
# Checks the couplings that must move together and that nothing else enforces.
#
# Each of these is a pair of values in different files that CI would happily let
# diverge: the build would succeed, the tests would pass, and the artefact would
# be wrong.
set -uo pipefail
cd "$(dirname "$0")/.." || exit 1

fail=0

# check <description> <a> <b>
#
# NOTE: an empty value is a failure, not a match. sed exits 0 when a pattern
# simply does not match, so a reformat that breaks BOTH extractions of a pair
# would otherwise compare "" against "" and report a pass - the check silently
# stops checking. (`set -e` does not help here for the same reason.)
check() {
    if [[ -z "$2" || -z "$3" ]]; then
        printf '  BROKEN %s: could not extract a value (%s vs %s)\n' "$1" "${2:-<empty>}" "${3:-<empty>}"
        fail=1
    elif [[ "$2" == "$3" ]]; then
        printf '  ok    %s (%s)\n' "$1" "$2"
    else
        printf '  DRIFT %s: %s != %s\n' "$1" "$2" "$3"
        fail=1
    fi
}

# version.yml bumps these three with three separate seds; a no-op sed is silent.
version_go="$(sed -n 's/^const Version string = "\(.*\)"/\1/p' main.go)"
version_docker="$(sed -n 's/.*org.opencontainers.image.version="\([^"]*\)".*/\1/p' Dockerfile)"
version_npm="$(sed -n 's/^  "version": "\([^"]*\)",/\1/p' package.json)"
check "version constant vs Dockerfile label" "${version_go}" "${version_docker}"
check "version constant vs package.json" "${version_go}" "${version_npm}"

go_mod="$(sed -n 's/^go \(.*\)/\1/p' go.mod)"
go_docker="$(sed -n '1s/FROM golang:\([^ ]*\).*/\1/p' Dockerfile)"
check "go.mod toolchain vs Dockerfile base" "${go_mod}" "${go_docker}"

lint_make="$(sed -n 's/^GOLANGCI_LINT_VERSION := \(.*\)/\1/p' Makefile)"
lint_ci="$(sed -n 's/^ *version: \(v2\..*\)/\1/p' .github/workflows/test.yml)"
check "golangci-lint pin: Makefile vs CI" "${lint_make}" "${lint_ci}"

# A setting removed in v7 must not still be taught in the docs - copying such an
# example into a workflow is now a hard failure.
stale=0
for name in $(sed -n 's/^\t"\([A-Z_]*\)":.*/\1/p' config.go); do
    if grep -rq "^[[:space:]]*${name}:" docs/example.md README.md 2>/dev/null; then
        printf '  DRIFT docs still configure the removed setting %s\n' "${name}"
        fail=1
        stale=1
    fi
done
[[ ${stale} -eq 0 ]] && printf '  ok    docs configure no removed settings\n'

# setup-go matches a toolchain directive BEFORE the go line, so one here would
# silently override go-version-file in every workflow.
if grep -q '^toolchain ' go.mod; then
    printf '  DRIFT go.mod has a toolchain directive; setup-go would prefer it over the go line\n'
    fail=1
else
    printf '  ok    go.mod has no toolchain directive\n'
fi

exit ${fail}
