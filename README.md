# git-release
[![Release](https://img.shields.io/github/release/golang-standards/project-layout.svg?style=flat-square)](https://github.com/anton-yurchenko/git-release/releases/latest)  

A GitHub Action for creating a new GitHub Release with Assets and Changelog whenever a Version Tag is pushed to the project.  

![PIC](docs/images/release.png)

## Features:
- Parse Tag to match Semantic Versioning (for example: 2.8.1).  
- Upload build artifacts (assets) to the release.  
- Add changelog to the release.  

## Manual:
1. Add your changes to **CHANGELOG.md** in the following format (according to [keepachangelog.com](https://keepachangelog.com/ "Keep a ChangeLog")):
```
## [2.1.1] - 2019-09-29
### Added
- Feature 1.
- Feature 2.

### Changed
- Logger format.

### Removed
- Old library.
- Configuration file.
```
2. Tag a commit with Version (according to [semver.org](https://semver.org/ "Semantic Versioning")).
    - Extensions like **alpha/beta/rc/...** are not supported.
3. Push and watch **Git-Release** publishing a Release on GitHub ;-)  
![PIC](docs/images/log.png)

## Configuration:
1. Create a GitHub **Personal Access Token** with **Full control of private repositories** permission, add it as a Secret to target repository.
2. Change the workflow to be triggered on Tag Push:
```
on:
  push:
    tags:
    - '*'
```
3. Add Release stage to your workflow:
   - Provide GITHUB_TOKEN from step 1.
   - DRAFT_RELEASE and PRE_RELEASE are optional, assumed **false** if not set.
   - Provide a list of assets as **args**
```
    - name: Release
      uses: anton-yurchenko/git-release@master
      env:
        GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
        DRAFT_RELEASE: "false"
        PRE_RELEASE: "false"
      with:
        args: |
          build/release/artifact-darwin-amd64.zip
          build/release/artifact-linux-amd64.zip
          build/release/artifact-windows-amd64.zip
```