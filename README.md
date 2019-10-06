# git-release
[![License](https://img.shields.io/github/license/anton-yurchenko/git-release?style=flat-square)](LICENSE.md) [![Release](https://img.shields.io/github/v/release/anton-yurchenko/git-release?style=flat-square)](https://github.com/anton-yurchenko/git-release/releases/latest) [![Docker Build](https://img.shields.io/docker/cloud/build/antonyurchenko/git-release?style=flat-square)](https://hub.docker.com/r/antonyurchenko/git-release) [![Docker Pulls](https://img.shields.io/docker/pulls/antonyurchenko/git-release?style=flat-square)](https://hub.docker.com/r/antonyurchenko/git-release)

A GitHub Action for creating a GitHub Release with Assets and Changelog whenever a Version Tag is pushed to the project.  

![PIC](docs/images/release.png)

## Features:
- Parse Tag to match Semantic Versioning.  
- Upload build artifacts (assets) to the release.  
- Add a changelog to the release.  

## Manual:
1. Add your changes to **CHANGELOG.md** in the following format (according to [keepachangelog.com](https://keepachangelog.com/en/1.0.0/ "Keep a ChangeLog")):
```
## [2.1.5] - 2019-10-01
### Added
- Feature 1.
- Feature 2.

### Changed
- Logger timestamp.

### Removed
- Old library.
- Configuration file.
```
2. Tag a commit with Version (according to [semver.org](https://semver.org/ "Semantic Versioning")).
    - Extensions like **alpha/beta/rc/...** are not supported.
3. Push and watch **Git-Release** publishing a Release on GitHub ;-)  
![PIC](docs/images/log.png)

## Configuration:
1. Change the workflow to be triggered on Tag Push:
    - Use either `'*'` or `'v*`
```
on:
  push:
    tags:
    - 'v*'
```
2. Add Release stage to your workflow:
    - **Optional**: Provide a list of assets as **args**
    - **Optional**: `DRAFT_RELEASE: "true"/"false"` - Save release as draft instead of publishing it (default `false`).
    - **Optional**: `PRE_RELEASE: "true"/"false"` - GitHub will point out that this release is identified as non-production ready (default `false`). 
    - **Optional**: `CHANGELOG_FILE: "changes.md"` - Changelog filename (default `CHANGELOG.md`).
```
    - name: Release
      uses: docker://antonyurchenko/git-release:latest
      env:
        GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
        DRAFT_RELEASE: "false"
        PRE_RELEASE: "false"
        CHANGELOG_FILE: "CHANGELOG.md"
      with:
        args: |
          build/release/artifact-darwin-amd64.zip
          build/release/artifact-linux-amd64.zip
          build/release/artifact-windows-amd64.zip
```

## Remarks:
- This action is automatically built at Docker Hub, and tagged with `latest / v1 / v1.0 / v1.0.0`. You may lock to a certain version instead of using **latest**. (*Recommended to lock against a major version, for example* `v1`).
- Instead of using pre-built image, you may build it during the execution of your flow by changing `docker://antonyurchenko/git-release:latest` to `anton-yurchenko/git-release@master`

## License
[MIT](LICENSE.md) © 2019-present Anton Yurchenko