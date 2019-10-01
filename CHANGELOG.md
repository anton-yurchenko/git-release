
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