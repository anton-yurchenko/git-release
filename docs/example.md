# Examples

- You may pass data between steps in a workflow using [environmental variables](https://help.github.com/en/actions/automating-your-workflow-with-github-actions/development-tools-for-github-actions#set-an-environment-variable-set-env)
- Some examples are based on `run` instruction, which may be easily replaces with another **GitHub Action**

## SemVer tag

![PIC](images/release.png)

<details><summary>Workflow</summary>

```yaml
name: release

on:
  push:
    tags:
      - "*"

jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - name: Checkout
        uses: actions/checkout@v7

      - name: Release
        uses: docker://antonyurchenko/git-release:latest
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
        with:
          assets: |
            darwin-amd64.zip
            linux-amd64.zip
            windows-amd64.zip
```

</details>

## Release title with prefix

![PIC](images/example-prefix.png)

<details><summary>Workflow</summary>

```yaml
name: release

on:
  push:
    tags:
      - "*"

jobs:
  release:
    runs-on: ubuntu-latest
    steps:
      - name: Checkout
        uses: actions/checkout@v7

      - name: Release
        uses: docker://antonyurchenko/git-release:latest
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
          RELEASE_NAME_PREFIX: "Release: "
        with:
          assets: |
            darwin-amd64.zip
            linux-amd64.zip
            windows-amd64.zip
```

</details>

## Release title with suffix

![PIC](images/example-suffix.png)

<details><summary>Workflow</summary>

```yaml
name: release

on:
  push:
    tags:
      - "*"

jobs:
  release:
    runs-on: ubuntu-latest
    steps:
      - name: Checkout
        uses: actions/checkout@v7

      - name: Release
        uses: docker://antonyurchenko/git-release:latest
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
          PRE_RELEASE: "true"
          RELEASE_NAME_SUFFIX: " (nightly build)"
        with:
          assets: |
            darwin-amd64.zip
            linux-amd64.zip
            windows-amd64.zip
```

</details>

## Release title with prefix and suffix

![PIC](images/example-prefix-suffix.png)

<details><summary>Workflow</summary>

Can be set as global environmental variables or provided directly to the action

```yaml
name: release

on:
  push:
    tags:
      - "*"

jobs:
  release:
    runs-on: ubuntu-latest
    steps:
      - name: Checkout
        uses: actions/checkout@v7

      - run: |
          export PREFIX="Release: "
          export SUFFIX=" (Codename: 'Ragnarok')"
          echo "::set-env name=RELEASE_NAME_PREFIX::$PREFIX"
          echo "::set-env name=RELEASE_NAME_SUFFIX::$SUFFIX"

      - name: Release
        uses: docker://antonyurchenko/git-release:latest
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
        with:
          assets: |
            darwin-amd64.zip
            linux-amd64.zip
            windows-amd64.zip
```

</details>

## Release title with different changelog file

![PIC](images/example-name.png)

<details><summary>Workflow</summary>

Can be set as global environmental variable or provided directly to the action

```yaml
name: release

on:
  push:
    tags:
      - "*"

jobs:
  release:
    runs-on: ubuntu-latest
    steps:
      - name: Checkout
        uses: actions/checkout@v7

      - run: |
          export TEXT="Release X"
          echo "::set-env name=RELEASE_NAME::$TEXT"

      - name: Release
        uses: docker://antonyurchenko/git-release:latest
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
          CHANGELOG_FILE: "CHANGES.md"
          ALLOW_EMPTY_CHANGELOG: "true"
        with:
          assets: |
            darwin-amd64.zip
            linux-amd64.zip
            windows-amd64.zip
```

</details>

## Release name and body templates

By default the release title is the tag and the body is the changelog section of the released version.
`NAME_TEMPLATE` and `BODY_TEMPLATE` replace them with [Go templates](https://pkg.go.dev/text/template) rendered with
the following fields:

| Field | Description |
|:-----:|:-----------:|
| `.Tag` | Git tag, for example `v1.2.3` |
| `.Version` | Semantic version, for example `1.2.3` (or `Unreleased`) |
| `.Major` `.Minor` `.Patch` | Semantic version components |
| `.Prerelease` | Semantic version pre-release identifier, for example `rc.1` |
| `.Owner` | Repository owner |
| `.Repo` | Repository name |
| `.CommitHash` | Commit the release points to |
| `.IsDraft` | `true` when the release is a draft |
| `.IsPreRelease` | `true` when the release is a pre-release |
| `.IsUnreleased` | `true` for a rolling `unreleased` release |
| `.Changelog` | Changelog entries of the released version. **`BODY_TEMPLATE` only** |
| `.Name` | The rendered release title. **`BODY_TEMPLATE` only** |

The name is rendered before the changelog is read, so `.Changelog` and `.Name` are available to the body template only.

A reference to a field that does not exist terminates the action instead of silently rendering an empty value.

:information_source: `{{ ... }}` here is a **Go** template, not a GitHub Actions `${{ ... }}` expression - GitHub does not interpolate it.

<details><summary>Workflow</summary>

```yaml
name: release

on:
  push:
    tags:
      - "*"

jobs:
  release:
    runs-on: ubuntu-latest
    steps:
      - name: Checkout
        uses: actions/checkout@v7

      - name: Release
        uses: docker://antonyurchenko/git-release:latest
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
          NAME_TEMPLATE: "Release {{ .Tag }}{{ if .IsPreRelease }} (pre-release){{ end }}"
          BODY_TEMPLATE: |
            ## {{ .Name }}

            {{ .Changelog }}

            ---
            Released from `{{ .CommitHash }}` in `{{ .Owner }}/{{ .Repo }}`
        with:
          assets: build/*.zip
```

</details>

## Asset Filename Pattern Matching

![PIC](images/release.png)

<details><summary>Workflow</summary>

```yaml
name: release

on:
  push:
    tags:
      - "*"

jobs:
  release:
    runs-on: ubuntu-latest
    steps:
      - name: Checkout
        uses: actions/checkout@v7

      - name: Release
        uses: docker://antonyurchenko/git-release:latest
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
        with:
          assets: build/*.zip
```

</details>

## Windows Runner

![PIC](images/release.png)

<details><summary>Workflow</summary>

```yaml
name: release

on:
  push:
    tags:
      - "*"

jobs:
  release:
    runs-on: windows-latest
    steps:
      - name: Checkout
        uses: actions/checkout@v7

      - name: Release
        uses: anton-yurchenko/git-release@master
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
        with:
          assets: |
            darwin-amd64.zip
            linux-amd64.zip
            windows-amd64.zip
```

</details>

## Unreleased

This will recreate a single released on each execution by deleting the previous release and creating a new one.
Changelog will be extracted from an `Unreleased` scope inside a CHANGELOG.md file.

Because this is an *"Unreleased"* release, it will always be marked as a **pre-release**.

`latest` tag will be used by default, this means that it will be moved with each execution and point to a different commit.

![PIC](images/unreleased.gif)

<details><summary>Workflow</summary>

```yaml
name: release

on:
  push:
    branches:
      - master

jobs:
  release:
    runs-on: ubuntu-latest
    steps:
      - name: Checkout
        uses: actions/checkout@v7

      - name: Release
        uses: docker://antonyurchenko/git-release:latest
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
          UNRELEASED: "update"
        with:
          assets: linux-amd64
```

</details>

## Unreleased with custom Tag

Identical to [Unreleased](#unreleased) but with a different git tag. (useful when `latest` tag is used for something else)

![PIC](images/unreleased-tag.png)

<details><summary>Workflow</summary>

```yaml
name: release

on:
  push:
    branches:
      - master

jobs:
  release:
    runs-on: ubuntu-latest
    steps:
      - name: Checkout
        uses: actions/checkout@v7

      - name: Release
        uses: docker://antonyurchenko/git-release:latest
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
          UNRELEASED: "update"
          UNRELEASED_TAG: future
        with:
          assets: linux-amd64
```

</details>
