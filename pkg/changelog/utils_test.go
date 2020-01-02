package changelog_test

import (
	"io"
	"testing"

	"github.com/spf13/afero"
)

const content string = `## [Unreleased]
- Unrelease feature.
- Parsing bug fixed.

## [1.0.1-beta] - 2018-01-28
### Added
- New feature.

### Fixed
- Fixed env.var bug.

## [1.0.0] - 2018-01-01
- First stable release.

## [0.3.0] - 2017-12-31
### Fixed
- Wrong message on success.
- Proper log message.

## [0.2.0] - 2016-10-01
### Added
- File reader.
- License.

### Changed
- Remove 'v' from release name.

### Fixed
- Create release without assets.

### Removed
- 'DRAFT_RELEASE=false' warning logging.
- 'PRE_RELEASE=false' warning logging.

## [0.1.0] - 2019-09-29
### Added
- Create GitHub Release.
- Upload Assets.
- Control Release Draft through env.var 'DRAFT_RELEASE'.
- Control Release Pre Release through env.var 'PRE_RELEASE'.

## [1.2.3----RC-SNAPSHOT.12.9.1--.12+788] - 2018-01-01
- Test valid builds.

## [2.0.0+build.1848] - 2018-01-01
- Test valid builds.

[Unreleased]: [0.3.0]: https://github.com/anton-yurchenko/git-release/compare/v0.2.0...v0.3.0
[0.2.0]: https://github.com/anton-yurchenko/git-release/compare/v0.1.0...v0.2.0
[0.1.0]: https://github.com/anton-yurchenko/git-release/releases/tag/v0.1.0`

var releasesContentMargins map[string]map[string]int = map[string]map[string]int{
	"1.0.1-beta": map[string]int{
		"start": 5,
		"end":   10,
	},
	"1.0.0": map[string]int{
		"start": 12,
		"end":   13,
	},
	"0.3.0": map[string]int{
		"start": 15,
		"end":   18,
	},
	"0.2.0": map[string]int{
		"start": 20,
		"end":   33,
	},
	"0.1.0": map[string]int{
		"start": 35,
		"end":   40,
	},
	"1.2.3----RC-SNAPSHOT.12.9.1--.12+788": map[string]int{
		"start": 42,
		"end":   43,
	},
	"2.0.0+build.1848": map[string]int{
		"start": 45,
		"end":   46,
	},
}

var releasesContent map[string]string = map[string]string{
	"1.0.1-beta": `### Added
- New feature.

### Fixed
- Fixed env.var bug.`,
	"1.0.0": `- First stable release.`,
	"0.3.0": `### Fixed
- Wrong message on success.
- Proper log message.`,
	"0.2.0": `### Added
- File reader.
- License.

### Changed
- Remove 'v' from release name.

### Fixed
- Create release without assets.

### Removed
- 'DRAFT_RELEASE=false' warning logging.
- 'PRE_RELEASE=false' warning logging.`,
	"0.1.0": `### Added
- Create GitHub Release.
- Upload Assets.
- Control Release Draft through env.var 'DRAFT_RELEASE'.
- Control Release Pre Release through env.var 'PRE_RELEASE'.`,
	"1.2.3----RC-SNAPSHOT.12.9.1--.12+788": `- Test valid builds.`,
	"2.0.0+build.1848":                     `- Test valid builds.`,
}

func createChangelog(fs afero.Fs, t *testing.T) string {
	file, err := fs.Create("CHANGELOG.md")
	if err != nil {
		t.Fatal("error creating CHANGELOG.md", err)
	}

	_, err = io.WriteString(file, content)
	if err != nil {
		t.Fatal("error writing to CHANGELOG.md", err)
	}

	return file.Name()
}
