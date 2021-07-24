# git-release

[![Release](https://img.shields.io/github/v/release/anton-yurchenko/git-release)](https://github.com/anton-yurchenko/git-release/releases/latest)
[![Code Coverage](https://codecov.io/gh/anton-yurchenko/git-release/branch/main/graph/badge.svg)](https://codecov.io/gh/anton-yurchenko/git-release)
[![Go Report Card](https://goreportcard.com/badge/github.com/anton-yurchenko/git-release)](https://goreportcard.com/report/github.com/anton-yurchenko/git-release)
[![Release](https://github.com/anton-yurchenko/git-release/actions/workflows/release.yml/badge.svg)](https://github.com/anton-yurchenko/git-release/actions/workflows/release.yml)
[![Docker Pulls](https://img.shields.io/docker/pulls/antonyurchenko/git-release)](https://hub.docker.com/r/antonyurchenko/git-release)
[![License](https://img.shields.io/github/license/anton-yurchenko/git-release)](LICENSE.md)

A **GitHub Action** for a **GitHub Release** creation with **Assets** and **Changelog** on new **Git Tag** in the repository.  

![PIC](docs/images/release.png)

## Features

- Parse Tag to match [Semantic Versioning](https://semver.org/)
- Upload build artifacts (assets) to the release
- Publish release with changelog
  - [Keep a Changelog](https://keepachangelog.com/) Compliant
  - [Common Changelog](https://common-changelog.org) Compliant
- Supported runners:
  - Linux AMD64
  - Linux ARM64
  - Windows
- Filename pattern matching
- Supports GitHub Enterprise
- Supports standard `v` prefix out of the box
- Allows custom SemVer prefixes
- Update a single pre-release with changes from Unreleased scope

## Manual

1. Add changes to `CHANGELOG.md`. *For example:*

    ```markdown
    ## [3.4.0] - 2020-07-10
    ### Added
    - Glob pattern support
    - Unit Tests
    - Log version
    
    ### Fixed
    - Exception on margins larger than context of changelog
    - Nil pointer exception in 'release' package
    
    ### Changed
    - Refactor JavaScript wrapper
    
    ## [3.3.0] - 2020-06-27
    ### Added
    - Wrapper script: allow execution on Windows runners
    
    ### Changed
    - Action execution through Git: from Docker to NodeJS
    
    [3.4.0]: https://github.com/anton-yurchenko/git-release/compare/v3.3.0...v3.4.0
    [3.3.0]: https://github.com/anton-yurchenko/git-release/releases/tag/v3.3.0
    ```

2. Tag a commit with Version (according to [semver.org](https://semver.org/ "Semantic Versioning"))
3. Push and watch **Git-Release** publishing a Release on GitHub :wink:

    ![PIC](docs/images/log.png)

## Configuration

1. Change the workflow to be triggered on new Tag:

    - For example `'*'` or a more specific like `'v*'`:

    ```yaml
    on:
      push:
        tags:
        - "v[0-9]+.[0-9]+.[0-9]+"
    ```

2. Add *Release* step to your workflow:

    ```yaml
        - name: Release
          uses: docker://antonyurchenko/git-release:latest
          env:
            GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
          with:
            args: build/*.zip
    ```

3. Configure *Release* step:

    - Specify release assets as action `arguments` (divided by one of: `new line`, `space`, `comma`, `pipe`)
    - Fine tune action configuration using environmental variables:

    | Environmental Variable  | Allowed Values | Default Value  | Description                                                                                                                |
    |:-----------------------:|:--------------:|:-----------------:|:--------------------------------------------------------------------------------------------------------------------------:|
    | `DRAFT_RELEASE`         | `true`/`false`    | `false`           | Publish a draft release                                                                                                    |
    | `PRE_RELEASE`           | `true`/`false`    | `false`           | Mark release non-production ready                                                                                          |
    | `CHANGELOG_FILE`        | `*`               | `CHANGELOG.md`    | Changelog filename (set `none` to silence a warning message if file does not exist)                                        |
    | `ALLOW_EMPTY_CHANGELOG` | `true`/`false`    | `false`           | Allow publishing a release without changelog                                                                               |
    | `TAG_PREFIX_REGEX`      | `*`               | `[v]?`            | Version tag prefix regex, for example `[a-z-]*` in order to parse `prerelease-1.1.0`                                       |
    | `RELEASE_NAME`          | `*`               | ""                | Complete release title (should not be combined with `RELEASE_NAME_PREFIX` and `RELEASE_NAME_SUFFIX`)                       |
    | `RELEASE_NAME_PREFIX`   | `*`               | ""                | Release title prefix                                                                                                       |
    | `RELEASE_NAME_SUFFIX`   | `*`               | ""                | Release title suffix                                                                                                       |
    | `UNRELEASED`            | `update`/`delete` | ""                | Set to `update` in order to allow deletion and recreation of the same release and its tag (intended to be used for `unreleased`/`latest` release only). Set to `delete` in order to delete a previously published `unreleased`/`latest` release.                                                                                     |
    | `UNRELEASED_TAG`        | `latest`       | `*`               | Use a custom tag for `unreleased`/`latest` release (tag will be created/deleted automatically)                             |

    *Configuration is provided as environmental variables (strings), so do not forget to enclose boolean values with quotes*

<details><summary>:information_source: Windows Runners</summary>

Execute **git-release** through JavaScrip Wrapper on Windows Runners.

```yaml
    - name: Release
      uses: anton-yurchenko/git-release@main
      env:
        GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
      with:
        args: |
            build\\darwin-amd64.zip
            build\\linux-amd64.zip
            build\\windows-amd64.zip
```

</details>

:information_source: [Configuration Examples](docs/example.md#examples)

## Remarks

- This action has multiple tags: `latest / v1 / v1.2 / v1.2.3`. You may lock to a certain version instead of using **latest**.  
(*Recommended to lock against a major version, for example* `v4`)
- Instead of using a pre-built Docker image, you may execute the action through JavaScript wrapper by changing `docker://antonyurchenko/git-release:latest` to `anton-yurchenko/git-release@main`
- `git-release` operates assets with pattern matching, this means that it is unable to validate whether an asset exists
- Docker image is published both to [**Docker Hub**](https://hub.docker.com/r/antonyurchenko/git-release) and [**GitHub Packages**](https://github.com/anton-yurchenko/git-release/packages). If you don't want to rely on **Docker Hub** but still want to use the dockerized action, you may switch from `uses: docker://antonyurchenko/git-release:latest` to `uses: docker://ghcr.io/anton-yurchenko/git-release:latest`
- Slashes (`/`) in asset filenames will be replaced with dashes (`-`)

## License

[MIT](LICENSE.md) © 2019-present Anton Yurchenko
