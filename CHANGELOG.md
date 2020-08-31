## [3.4.1] - 2020-08-31
### Changed
- Update Dependencies

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

## [3.2.0] - 2020-06-04
### Fixed
- Ignored `ALLOW_EMPTY_CHANGELOG=true` failed to create a release.

### Added
- `CHANGELOG_FILE=none` will skip changelog file validation. This allows to create a release from empty workspace.
- Upgrade GoLang to 1.14.4
- Upgrade dependencies

## [3.1.2] - 2020-04-10
### Fixed
- [Issue #16](https://github.com/anton-yurchenko/git-release/issues/16) - Error parsing tags with slashes. (*Thanks to [Jonathan Hilgart](https://github.com/jonhilgart22)*)
- Support `.` in organization and repository names.

## [3.1.1] - 2020-03-25
### Fixed
- [Issue #14](https://github.com/anton-yurchenko/git-release/issues/14) - Ignored first release link in a comment. (*Thanks to [Luiz Ferraz](https://github.com/Fryuni)*)

### Changed
- Upgrade GoLang to 1.14.1
- Upgrade dependencies

## [3.1.0] - 2020-02-17
### Added
- [Issue #10](https://github.com/anton-yurchenko/git-release/issues/10) - Release Title manipulation through `RELEASE_NAME`, `RELEASE_NAME_PREFIX`, `RELEASE_NAME_POSTFIX`. (*Thanks to [Victor](https://github.com/victoraugustolls) for suggesting a change*)

## [3.0.1] - 2020-01-08
### Fixed
- Empty release name

## [3.0.0] - 2020-01-05
This is a major release because of a certain behavior change:  
- *Tag (without prefix) should be identical to Changelog Version in order for changes to be mapped (for example tag `v3.0.0-rc.1` is expected to be listed as `3.0.0-rc.1` in changelog).*
- *By default valid semver version is expected. Prefix should be explicitly allowed by enabling `ALLOW_TAG_PREFIX`*

### Changed
- Better `GITHUB_REPOSITORY` regex validation
- Improved **Changelog** package parsing capabilities
- Tag should match Changelog Version (excluding prefix)

### Fixed
- Semantic Versioning compliance
- Keep a Changelog compliance

### Added
- `ALLOW_TAG_PREFIX` to control version prefix like `v` or `release`

## [2.0.2] - 2019-12-29
### Added
- CircleCI integrated as a Continuous Integration system
- GolangCI integrated as a Continuous Code Quality system
- CodeCov integrated as a Continuous Code Quality system

### Changed
- DockerHub setup as a Continuous Delivery system

## [2.0.1] - 2019-12-28
### Changed
- Disable unit testing on Docker Hub auto builds

## [2.0.0] - 2019-12-28
This is a major release as most of the code was refactored and some behavior was changed, for example "Tag version is set as a release title".

### Fixed
- Artifact files not found caused panic - all files now being validated before release creation
- Custom changelog file now being validated before release creation
- Arguments parsing fixed

### Added
- Unit testing
- Docker image now built from scratch, resulting in decreased size 139.73MB -> 2.43MB, improving action overall speed.
- **app** package
- `ALLOW_EMPTY_CHANGELOG` env.var to allow publishing a release without changelog (default **false**)
- Artifacts (provided as arguments) can now be separated by one of: `new line '\n', pipe '|', space ' ', comma ','`

### Changed
- **local** package renamed to **repository**
- **remote** package splitted into 2 packages: **asset**, **release**
- Tag version is set as a release title

## [1.1.0] - 2019-12-21
### Added
- [PR #3](https://github.com/anton-yurchenko/git-release/pull/3) - Allow any prefix to semver tag. (*Thanks to [Taylor Becker](https://github.com/tajobe) for the PR*)

### Fixed
- [PR #3](https://github.com/anton-yurchenko/git-release/pull/3) - PreRelease overwriting Draft configuration. (*Thanks to [Taylor Becker](https://github.com/tajobe) for reporting an issue*)

## [1.0.0] - 2019-10-01
- First stable release.

## [0.2.1] - 2019-10-01
### Fixed
- Wrong PRE_RELEASE message when set.
- Correct 'creating release' log message.

## [0.2.0] - 2019-10-01
### Added
- Changelog reader.
- MIT License.

### Changed
- Remove 'v' from release name.

### Fixed
- Create release without assets.

## [0.1.1] - 2019-09-29
### Added
- Tag regex to match v1.0.0 and 1.0.0.
- Log when procedure finished.

### Removed
- 'DRAFT_RELEASE=false' warning logging.
- 'PRE_RELEASE=false' warning logging.

## [0.1.0] - 2019-09-29
### Added
- Create GitHub Release.
- Upload Assets.
- Control Release Draft through env.var 'DRAFT_RELEASE'.
- Control Release Pre Release through env.var 'PRE_RELEASE'.