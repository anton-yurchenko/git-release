# Changelog

## [Unreleased]

:warning: GitHub Actions initiate a deprecation process for [Node16](https://github.blog/changelog/2023-09-22-github-actions-transitioning-from-node-16-to-node-20/)

### Fixed

- [Issue #122](https://github.com/anton-yurchenko/git-release/issues/122) Latest pre-release always recreated as Draft (*Thanks to [Taylor Becker](https://github.com/tajobe)*)

### Changed

- Update dependencies
- Update Golang version to v1.21.5
- **Breaking:** Update NodeJS version to v20.11.0

## [5.0.2] - 2023-04-09

### Fixed

- [Issue #90](https://github.com/anton-yurchenko/git-release/issues/90) Awkward behavior in UNRELEASED flow when tag is present without release itself (*Thanks to [Benjamin K.](https://github.com/treee111)*)

### Changed

- Update dependencies
- Update Golang version

## [5.0.1] - 2022-12-14

### Fixed

- [Issue #86](https://github.com/anton-yurchenko/git-release/issues/86) Panic on empty unreleased changelog (*Thanks to [Taylor Becker](https://github.com/tajobe)*)

### Changed

- Update dependencies
- Update Golang version

## [5.0.0] - 2022-09-14

### Changed

- Update dependencies
- Update Golang version
- Update JavaScript version

## [4.2.4] - 2022-02-20

### Fixed

- [Issue #64](https://github.com/anton-yurchenko/git-release/issues/64) Panic on missing API response (*Thanks to [rgriebl](https://github.com/rgriebl)*)

## [4.2.3] - 2022-02-20

### Changed

- Keep retrying assets upload even if `422/UnprocessableEntity`/`502/BadGateway` is encountered and asset was not found on the partially created release

### Fixed

- [Issue #64](https://github.com/anton-yurchenko/git-release/issues/64) Panic on missing API response (*Thanks to [rgriebl](https://github.com/rgriebl)*)

## [4.2.2] - 2022-02-05

### Fixed

- [Issue #61](https://github.com/anton-yurchenko/git-release/issues/61) Improve assets upload retry mechanism (*Thanks to [kongsgard](https://github.com/kongsgard), [rgriebl](https://github.com/rgriebl)*)

### Added

- [Issue #61](https://github.com/anton-yurchenko/git-release/issues/61) Recover from 422/UnprocessableEntity and 502/BadGateway errors during assets upload (*Thanks to [kongsgard](https://github.com/kongsgard), [rgriebl](https://github.com/rgriebl)*)

### Changed

- Update dependencies

## [4.2.1] - 2021-12-15

### Fixed

- Error uploading an asset during retry loop

## [4.2.0] - 2021-11-25

### Added

- Retry assets uploads

### Changed

- Update dependencies

## [4.1.2] - 2021-10-18

### Fixed

- [Issue #54](https://github.com/anton-yurchenko/git-release/issues/54) Empty scopes during changelog parsing (*Thanks to [Wolf2323](https://github.com/Wolf2323)*)

### Changed

- Update dependencies

## [4.1.1] - 2021-08-22

### Fixed

- Crash when changelog does not contain changes for a version
- Changelog references

### Changed

- Update to GoLang 1.17
- Update dependencies

## [4.1.0] - 2021-07-02

### Added

- [Issue #47](https://github.com/anton-yurchenko/git-release/issues/47) Recreate `Unreleased` release on each execution (*Thanks to [cb80](https://github.com/cb80)*)

## [4.0.1] - 2021-06-24

### Changed

- Provide more descriptive error messages

## [4.0.0] - 2021-06-16

### Changed

- Enforce changelog file format to comply with **Keep a Changelog**/**Common Changelog**
- Allow `v` prefix without `ALOW_TAG_PREFIX` (*still required for other prefixes*)
- Update Dependencies
- Update to Golang 1.16
- Not existing changelog file won't fail the execution, but will log this as error. Set `CHANGELOG_FILE=none` to silence an error message

### Added

- [Issue #46](https://github.com/anton-yurchenko/git-release/issues/46) Support GitHub Enterprise (*Thanks to [cb80](https://github.com/cb80)*)

### Removed

- `ALLOW_TAG_PREFIX` was replaced with `TAG_PREFIX_REGEX`
- `RELEASE_NAME_POSTFIX` was replaced with `RELEASE_NAME_SUFFIX`
- Logs on empty (default) variables: [`DRAFT_RELEASE`, `PRE_RELEASE`, `ALLOW_EMPTY_CHANGELOG`, `RELEASE_NAME`, `RELEASE_NAME_PREFIX`, `RELEASE_NAME_SUFFIX`, `CHANGELOG_FILE`]

### Fixed

- Version extraction
- Custom prefix matching
- Theoretically possible incomplete assets upload
- Nil pointer reference on empty release

## [3.5.0] - 2021-05-01

### Changed

- Update Dependencies
- Build project on GitHub Actions

### Added

- [Issue #44](https://github.com/anton-yurchenko/git-release/issues/44) - Support ARM64 by building a multi-arch docker image (*Thanks to [rsliotta](https://github.com/rsliotta)*)

## [3.4.4] - 2021-03-13

### Changed

- Update Dependencies

### Deprecated

- `RELEASE_NAME_POSTFIX` will be changed to `RELEASE_NAME_SUFFIX` in the next release

## [3.4.3] - 2021-01-02

### Fixed

- [PR #38](https://github.com/anton-yurchenko/git-release/pull/38) - Version prefix greedy quantifier caused incorrect parsing of major version higher then `9`. (*Thanks to [rgriebl](https://github.com/rgriebl)*)

### Changed

- Update Dependencies

## [3.4.2] - 2020-10-25

### Changed

- Update Dependencies
- Make `version` output more specific

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

- Ignored `ALLOW_EMPTY_CHANGELOG=true` failed to create a release

### Added

- `CHANGELOG_FILE=none` will skip changelog file validation. This allows to create a release from empty workspace

### Changed

- Upgrade GoLang to 1.14.4
- Upgrade dependencies

## [3.1.2] - 2020-04-10

### Fixed

- [Issue #16](https://github.com/anton-yurchenko/git-release/issues/16) - Error parsing tags with slashes. (*Thanks to [Jonathan Hilgart](https://github.com/jonhilgart22)*)
- Support `.` in organization and repository names

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

This is a major release as most of the code was refactored and some behavior was changed, for example "Tag version is set as a release title"

### Fixed

- Artifact files not found caused panic - all files now being validated before release creation
- Custom changelog file now being validated before release creation
- Arguments parsing fixed

### Added

- Unit testing
- Docker image now built from scratch, resulting in decreased size 139.73MB -> 2.43MB, improving action overall speed
- **app** package
- `ALLOW_EMPTY_CHANGELOG` env.var to allow publishing a release without changelog (default **false**)
- Artifacts (provided as arguments) can now be separated by one of: `new line '\n', pipe '|', space ' ', comma ','`

### Changed

- **local** package renamed to **repository**
- **remote** package split into 2 packages: **asset**, **release**
- Tag version is set as a release title

## [1.1.0] - 2019-12-21

### Added

- [PR #3](https://github.com/anton-yurchenko/git-release/pull/3) - Allow any prefix to semver tag. (*Thanks to [Taylor Becker](https://github.com/tajobe) for the PR*)

### Fixed

- [PR #3](https://github.com/anton-yurchenko/git-release/pull/3) - PreRelease overwriting Draft configuration. (*Thanks to [Taylor Becker](https://github.com/tajobe) for reporting an issue*)

## [1.0.0] - 2019-10-01

- First stable release

## [0.2.1] - 2019-10-01

### Fixed

- Wrong PRE_RELEASE message when set
- Correct `creating release` log message

## [0.2.0] - 2019-10-01

### Added

- Changelog reader
- MIT License

### Changed

- Remove `v` from release name

### Fixed

- Create release without assets

## [0.1.1] - 2019-09-29

### Added

- Tag regex to match v1.0.0 and 1.0.0
- Log when procedure finished

### Removed

- `DRAFT_RELEASE=false` warning logging
- `PRE_RELEASE=false` warning logging

## [0.1.0] - 2019-09-29

### Added

- Create GitHub Release
- Upload Assets
- Control Release Draft through env.var `DRAFT_RELEASE`
- Control Release Pre Release through env.var `PRE_RELEASE`

[Unreleased]: https://github.com/anton-yurchenko/git-release/compare/v5.0.2...HEAD
[5.0.2]: https://github.com/anton-yurchenko/git-release/compare/v5.0.1...v5.0.2
[5.0.1]: https://github.com/anton-yurchenko/git-release/compare/v5.0.0...v5.0.1
[5.0.0]: https://github.com/anton-yurchenko/git-release/compare/v4.2.4...v5.0.0
[4.2.4]: https://github.com/anton-yurchenko/git-release/compare/v4.2.3...v4.2.4
[4.2.3]: https://github.com/anton-yurchenko/git-release/compare/v4.2.2...v4.2.3
[4.2.2]: https://github.com/anton-yurchenko/git-release/compare/v4.2.1...v4.2.2
[4.2.1]: https://github.com/anton-yurchenko/git-release/compare/v4.2.0...v4.2.1
[4.2.0]: https://github.com/anton-yurchenko/git-release/compare/v4.1.2...v4.2.0
[4.1.2]: https://github.com/anton-yurchenko/git-release/compare/v4.1.1...v4.1.2
[4.1.1]: https://github.com/anton-yurchenko/git-release/compare/v4.1.0...v4.1.1
[4.1.0]: https://github.com/anton-yurchenko/git-release/compare/v4.0.1...v4.1.0
[4.0.1]: https://github.com/anton-yurchenko/git-release/compare/v4.0.0...v4.0.1
[4.0.0]: https://github.com/anton-yurchenko/git-release/compare/v3.5.0...v4.0.0
[3.5.0]: https://github.com/anton-yurchenko/git-release/compare/v3.4.4...v3.5.0
[3.4.4]: https://github.com/anton-yurchenko/git-release/compare/v3.4.3...v3.4.4
[3.4.3]: https://github.com/anton-yurchenko/git-release/compare/v3.4.2...v3.4.3
[3.4.2]: https://github.com/anton-yurchenko/git-release/compare/v3.4.1...v3.4.2
[3.4.1]: https://github.com/anton-yurchenko/git-release/compare/v3.4.0...v3.4.1
[3.4.0]: https://github.com/anton-yurchenko/git-release/compare/v3.3.0...v3.4.0
[3.3.0]: https://github.com/anton-yurchenko/git-release/compare/v3.2.0...v3.3.0
[3.2.0]: https://github.com/anton-yurchenko/git-release/compare/v3.1.2...v3.2.0
[3.1.2]: https://github.com/anton-yurchenko/git-release/compare/v3.1.1...v3.1.2
[3.1.1]: https://github.com/anton-yurchenko/git-release/compare/v3.1.0...v3.1.1
[3.1.0]: https://github.com/anton-yurchenko/git-release/compare/v3.0.1...v3.1.0
[3.0.1]: https://github.com/anton-yurchenko/git-release/compare/v3.0.0...v3.0.1
[3.0.0]: https://github.com/anton-yurchenko/git-release/compare/v2.0.2...v3.0.0
[2.0.2]: https://github.com/anton-yurchenko/git-release/compare/v2.0.1...v2.0.2
[2.0.1]: https://github.com/anton-yurchenko/git-release/compare/v2.0.0...v2.0.1
[2.0.0]: https://github.com/anton-yurchenko/git-release/compare/v1.1.0...v2.0.0
[1.1.0]: https://github.com/anton-yurchenko/git-release/compare/v1.0.0...v1.1.0
[1.0.0]: https://github.com/anton-yurchenko/git-release/compare/v0.2.1...v1.0.0
[0.2.1]: https://github.com/anton-yurchenko/git-release/compare/v0.2.0...v0.2.1
[0.2.0]: https://github.com/anton-yurchenko/git-release/compare/v0.1.1...v0.2.0
[0.1.1]: https://github.com/anton-yurchenko/git-release/compare/v0.0.1...v0.1.1
[0.1.0]: https://github.com/anton-yurchenko/git-release/releases/tag/v0.1.0
