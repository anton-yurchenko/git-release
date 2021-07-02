package main

import (
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/anton-yurchenko/git-release/release"
	changelog "github.com/anton-yurchenko/go-changelog"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/afero"
)

// Configuration is a git-release settings struct
type Configuration struct {
	AllowEmptyChangelog bool
	IgnoreChangelog     bool
	UnreleasedCreate    bool
	UnreleasedDelete    bool
	TagPrefix           string
	ReleaseName         string
	ReleaseNamePrefix   string
	ReleaseNameSuffix   string
	ChangelogFile       string
}

// GetConfig sets validated Release/Changelog configuration and returns github.com Token
func GetConfig(fs afero.Fs) (*Configuration, error) {
	conf := new(Configuration)

	if strings.ToLower(os.Getenv("ALLOW_EMPTY_CHANGELOG")) == "true" {
		conf.AllowEmptyChangelog = true
	}

	switch os.Getenv("UNRELEASED") {
	case "update":
		conf.UnreleasedCreate = true
	case "delete":
		conf.UnreleasedDelete = true
	case "":
		// do nothing
	default:
		return nil, errors.New("UNRELEASED not supported, possible values are [update, delete]")
	}

	conf.TagPrefix = os.Getenv("TAG_PREFIX_REGEX")
	conf.ReleaseName = os.Getenv("RELEASE_NAME")
	conf.ReleaseNamePrefix = os.Getenv("RELEASE_NAME_PREFIX")
	conf.ReleaseNameSuffix = os.Getenv("RELEASE_NAME_SUFFIX")

	if conf.ReleaseName != "" && ((conf.ReleaseNamePrefix != "" && conf.ReleaseNameSuffix != "") || (conf.ReleaseNamePrefix != "" || conf.ReleaseNameSuffix != "")) {
		return nil, errors.New("both RELEASE_NAME and RELEASE_NAME_PREFIX / RELEASE_NAME_SUFFIX are set (expected RELEASE_NAME or combination/one of RELEASE_NAME_PREFIX and RELEASE_NAME_SUFFIX)")
	}

	c := os.Getenv("CHANGELOG_FILE")
	if c == "" {
		c = "CHANGELOG.md"
	}
	conf.ChangelogFile = path.Join(os.Getenv("GITHUB_WORKSPACE"), c)

	b, err := afero.Exists(fs, conf.ChangelogFile)
	if err != nil {
		return nil, errors.Wrap(err, "error validating changelog file")
	}

	if !b {
		if c != "none" {
			log.Errorf("changelog file %v not found!", c)
		}

		conf.ChangelogFile = ""
		conf.IgnoreChangelog = true
	}

	// NOTE: deprecation warnings
	if os.Getenv("RELEASE_NAME_POSTFIX") != "" {
		log.Fatalf(`'RELEASE_NAME_POSTFIX' was deprecated.
- use 'RELEASE_NAME_SUFFIX' instead`)
	}
	if os.Getenv("ALLOW_TAG_PREFIX") != "" {
		log.Fatalf(`'ALLOW_TAG_PREFIX' was deprecated.
- in case your tag has a 'v' prefix, you can safely remove 'ALLOW_TAG_PREFIX' env.var
- if you have another prefix, provide a regex expression through 'TAG_PREFIX_REGEX' instead`)
	}

	return conf, nil
}

func (c *Configuration) GetChangelog(fs afero.Fs, rel *release.Release) (string, error) {
	p, err := changelog.NewParserWithFilesystem(fs, c.ChangelogFile)
	if err != nil {
		return "", errors.Wrap(err, "error loading changelog file")
	}

	changes, err := p.Parse()
	if err != nil {
		return "", errors.Wrap(err, "error parsing changelog file")
	}

	var msg string
	if rel.Reference.Version == "Unreleased" {
		if changes.Unreleased != nil {
			return changes.Unreleased.Changes.ToString(), nil
		} else {
			msg = "changelog file does not contain changes in Unreleased scope"
		}
	} else {
		r := changes.GetRelease(rel.Reference.Version)
		if r == nil {
			msg = fmt.Sprintf("no changes were found for version %v.", rel.Reference.Version) + ` make sure that:
- changelog file contains a required version
- version has changes
- changelog format is compliant with either 'Keep a Changelog' or 'Common Changelog'`
		} else {
			return r.Changes.ToString(), nil
		}
	}

	if msg != "" {
		if !c.AllowEmptyChangelog {
			return "", errors.New(msg)
		} else {
			log.Warn(msg)
		}
	}

	return "", nil
}
