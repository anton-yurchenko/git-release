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
        uses: docker://antonyurchenko/git-release:v7
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
        with:
          args: |
            darwin-amd64.zip
            linux-amd64.zip
            windows-amd64.zip
```

</details>

## Release title with a prefix

See [Release Name and Body Templates](../README.md#templates) for the available fields.

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
        uses: docker://antonyurchenko/git-release:v7
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
          NAME_TEMPLATE: "Release: {{ .Tag }}"
        with:
          args: |
            darwin-amd64.zip
            linux-amd64.zip
            windows-amd64.zip
```

</details>

## Release title with a suffix

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
        uses: docker://antonyurchenko/git-release:v7
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
          PRE_RELEASE: "true"
          NAME_TEMPLATE: "{{ .Tag }} (nightly build)"
        with:
          args: |
            darwin-amd64.zip
            linux-amd64.zip
            windows-amd64.zip
```

</details>

## Release title with a prefix and a suffix

![PIC](images/example-prefix-suffix.png)

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
        uses: docker://antonyurchenko/git-release:v7
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
          NAME_TEMPLATE: "Release: {{ .Tag }} (Codename: 'Ragnarok')"
        with:
          args: |
            darwin-amd64.zip
            linux-amd64.zip
            windows-amd64.zip
```

</details>

## Release title with different changelog file

![PIC](images/example-name.png)

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
        uses: docker://antonyurchenko/git-release:v7
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
          NAME_TEMPLATE: "Release X"
          CHANGELOG_FILE: "CHANGES.md"
          ALLOW_EMPTY_CHANGELOG: "true"
        with:
          args: |
            darwin-amd64.zip
            linux-amd64.zip
            windows-amd64.zip
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
        uses: docker://antonyurchenko/git-release:v7
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
        with:
          args: build/*.zip
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
        uses: anton-yurchenko/git-release@v7
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
        with:
          args: |
            darwin-amd64.zip
            linux-amd64.zip
            windows-amd64.zip
```

</details>

## Unreleased

This will recreate a single released on each execution by deleting the previous release and creating a new one.
Changelog will be extracted from an `Unreleased` scope inside a CHANGELOG.md file.

Because this is an _"Unreleased"_ release, it will always be marked as a **pre-release**.

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
        uses: docker://antonyurchenko/git-release:v7
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
          UNRELEASED: "update"
        with:
          args: linux-amd64
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
        uses: docker://antonyurchenko/git-release:v7
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
          UNRELEASED: "update"
          UNRELEASED_TAG: future
        with:
          args: linux-amd64
```

</details>
