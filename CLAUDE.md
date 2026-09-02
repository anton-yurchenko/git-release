# git-release

A GitHub Action that creates a GitHub Release with assets and a changelog when a tag is pushed.
Go binary, shipped two ways: a Docker image and a JavaScript wrapper.

## Commands

```sh
make tools          # install golangci-lint + mockery - REQUIRED before lint/test/mocks
make lint
make test           # lint + the full race suite (~3 min, see below)
make build          # lint + test + rebuild bin/ - REWRITES TRACKED FILES
make mocks          # regenerate mocks/ from .mockery.yml
```

The fast inner loop, because `make test` takes ~176s and 172s of that is one test:

```sh
go test -count=1 -race -skip '^TestUpload$' $(go list ./... | grep -v '/mocks$')   # ~6s
go build -o /dev/null ./...    # compile check without touching bin/
```

`TestUpload` really sleeps: `math.Pow(3, i+1)` gives 9s + 27s + 81s per retry-exhausting case
(`release/asset.go`). It cannot be sped up without adding a seam to production code. Run the full
suite before handing work over, not during iteration.

Tests in `package main` fail if your shell exports any git-release setting (`RELEASE_NAME*`,
`ALLOW_TAG_PREFIX`, `INPUT_ARGS`). If a test passes in CI and fails locally, check that first.

## Layout

| path | role |
|---|---|
| `main.go` | entrypoint; env gate, orchestration |
| `config.go` | settings, templates, the `removed` map |
| `env/` | environment reading, strict bool/enum parsing |
| `github.go` | authenticated go-github client |
| `release/` | the domain: reference parsing, publishing, asset upload |
| `mocks/` | generated - never hand-edit |

Module path is the bare name `git-release`, so imports are `git-release/release`, not a URL.

## Conventions

- Errors: `github.com/pkg/errors`, wrapped with lowercase context (`errors.Wrap(err, "error reading changelog")`).
  Every failure in `main()` ends at `log.Fatal`.
- Logging: `logrus`, aliased `log`, configured once in `main.go`'s `init()`. Output goes to **stdout**
  with **forced colours**, so anything grepping it must strip ANSI escapes first.
- Tests: table-driven, `suite := map[string]test{...}`, iterated with a counter and
  `t.Logf("Test Case %v/%v - %s", ...)`. Match the existing shape rather than introducing a new one.
- After changing an interface in `release/modal.go`, run `make mocks`.

## Adding a setting

Every setting is an environment variable. Touch, in order:

1. `config.go` (or `release/release.go`) - read it via `env.Get` / `env.Bool` / `env.Enum`
2. `README.md` - the configuration table
3. `docs/example.md` - if it deserves an example
4. `CHANGELOG.md` - one line under `## [Unreleased]`
5. `../git-release-dev/lib/case.sh` - **both** the `emit_env` translation and the `unset` list in
   `run_case`, plus a case under `cases/`

Missing step 5's `unset` leaks the value into every later case in the run.

**Do not declare it as an `action.yml` input.** `uses: docker://antonyurchenko/git-release:vN` is the
documented primary usage and never reads `action.yml` - no defaults, no validation, no deprecation
warnings. Declaring inputs buys marketplace metadata for the wrapper minority while costing everyone a
second spelling to learn and support. `args` is the sole exception; it predates this and is read from
`INPUT_ARGS`, which the runner sets in both modes.

**Never add a `default:` to an input.** A defaulted input is always exported on the wrapper path and
never on the container path, which silently forks the two distribution modes.

## Changelog

`CHANGELOG.md` is read by users, not machines. Under `## [Unreleased]`:

- One line per change - a PR, a feature, a fix. Not one line per commit.
- A short, user-meaningful sentence saying what changed. No details, no code snippets, no rationale.
  Anyone who needs more reads the PR.
- Keep a Changelog scopes: `Added`, `Changed`, `Deprecated`, `Removed`, `Fixed`, `Security`.
- Prefix a breaking change with `**Breaking:**`.
- Credit external contributors as `(*Thanks to [Name](https://github.com/handle)*)`.
- Leave everything in `## [Unreleased]`; the maintainer slices it at release time.

Two invariants that look like mistakes and are not:

- **The file ends with no trailing newline.** `changelog-version` rewrites it that way on every bump.
  An editor that adds one creates a phantom diff the next release strips again.
- **`## [Unreleased]` must never be empty.** `go-changelog` fails the version bump with
  `missing 'Unreleased' section`, after the release workflow has already rewritten versions and rebuilt
  binaries. Landing a change with no changelog entry blocks the next release.

Version headings must match `## [X.Y.Z] - YYYY-MM-DD` exactly, and section order in the published
release body is fixed by the library, not by the file.

## Releasing

Dispatch **Create New Version** on `main` with `vX.Y.Z`. It rewrites the three version strings, commits,
and pushes the tag. The tag push triggers `release.yml`, which tests, scans, builds the multi-arch images
and publishes the GitHub release.

**The tag is pushed before anything tests it.** `version.yml` only bumps and tags - every gate lives in
`release.yml`, downstream of the tag. A failed release therefore leaves a tag pointing at a commit with
nothing published: no GitHub release, no images, `:latest` unmoved. Recovery depends on what failed.

- **A workflow file, or the code** - delete the tag and re-cut it. Re-running a tag-push workflow loads
  the workflow file *from the tagged commit*, so a fix landed on `main` afterwards can never reach that
  tag; the re-run fails identically, forever.

      git push origin :refs/tags/vX.Y.Z && git tag -d vX.Y.Z
      # merge the fix, then dispatch Create New Version again with the same version

  Deleting is safe precisely because nothing published - and if the release job did not complete, nothing
  did. Only branch rulesets exist here; tags are unprotected.

- **A secret** - `gh run rerun <id> --failed`. Secrets resolve at job start, so the same commit picks up
  the new value on the next attempt. No tag surgery, and `test`/`scan` keep their results.

A green run is not proof, since both failure modes above are silent. Verify what actually landed:

    gh release view vX.Y.Z --json tagName,isDraft,isPrerelease
    curl -s "https://hub.docker.com/v2/repositories/antonyurchenko/git-release/tags/?page_size=5"
    TOK=$(curl -s "https://ghcr.io/token?scope=repository:anton-yurchenko/git-release:pull" | jq -r .token)
    curl -s -H "Authorization: Bearer $TOK" https://ghcr.io/v2/anton-yurchenko/git-release/tags/list
    docker run --rm antonyurchenko/git-release:vX.Y.Z   # DEBUG banner must read the new version

Three things rot between releases. All three broke v7.0.0, none is caught by CI:

- **Secrets expire while the repo is dormant.** `DOCKER_HUB_TOKEN` was five years old and failed with
  `unauthorized: incorrect username or password`. Before a release that follows a long gap, check ages
  with `gh api repos/anton-yurchenko/git-release/actions/secrets`.
- **The self-release pin is manual.** `release.yml` releases this action using
  `docker://antonyurchenko/git-release:vN`. Dependabot never touches `docker://`, so bump it by hand
  after each major.
- **Pinned actions can outgrow the module.** `Templum/govulncheck-action` ships its own Go and runs with
  `GOTOOLCHAIN=local`, so it could not build a `go 1.27` module. Any action that compiles this code needs
  `go-version-file: go.mod`, not a bundled toolchain.

Right after a release `## [Unreleased]` is empty, so the next **Create New Version** fails with
`missing 'Unreleased' section` until a change lands. That is the Changelog invariant above, seen from the
release side.

## Traps

- **`bin/` is tracked and no CI job rebuilds it.** It is what the JavaScript wrapper executes. Any Go
  change - including a dependency bump - leaves `uses: anton-yurchenko/git-release@vN` running the old
  binary while `docker://` runs the new one, all green. Run `make build` and commit `bin/`.
- **`node_modules/` is tracked** and Dependabot only rewrites `package.json`/`package-lock.json`. Before
  merging an npm PR: `npm ci && git add -A node_modules package-lock.json package.json`.
- **`.gitignore` has `release/*` + `!release/*.go`**, so anything under `release/` that is not a
  top-level `.go` file is silently untracked - git never descends into an excluded directory. A new
  sub-package or testdata fixture will pass locally and fail in CI.
- **`version.yml` owns three version strings** via brittle `sed`s: `main.go`'s `Version` const, the
  Dockerfile `LABEL`, and `package.json`'s `"version"` (which needs its trailing comma). Never hand-edit
  or reformat them - a no-op sed ships a release reporting the previous version.
- **`go.mod` says `go 1.27.0` with no `toolchain` line.** That is deliberate: `setup-go` matches a
  `toolchain` directive first and would override `go-version-file`. Never set `GOTOOLCHAIN=local` here.
- **`go install` builds a tool with the tool's own toolchain**, so a linter or mockery built with an
  older Go cannot type-check this module. The Makefile pins `GOTOOLCHAIN` for exactly this. If `make
  mocks` complains about invalid keys in `.mockery.yml`, the binary is v2 - check `mockery --version`.
- **`TestUpload` asserts weakly** - its error check is inside `if err != nil` and it never calls
  `AssertExpectations`. Tighten it before changing `Asset.Upload`, or the most expensive test in the
  suite will not catch your regression.

## Testing against the real API

Behaviour is verified in the sibling repo `anton-yurchenko/git-release-dev`, against the live GitHub
API - nothing is mocked at that level. Push to `dev` here and `dev-image.yml` publishes
`ghcr.io/anton-yurchenko/git-release:dev`; then run the `all` workflow there. See its `README.md`.
