package main

import (
	"os"
	"path"
	"strings"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/afero"
)

// Configuration is a git-release settings struct
type Configuration struct {
	AllowEmptyChangelog bool
	IgnoreChangelog     bool
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
		log.Fatalf("'RELEASE_NAME_POSTFIX' was deprecated, use 'RELEASE_NAME_SUFFIX' instead")
	}
	if os.Getenv("ALLOW_TAG_PREFIX") != "" {
		log.Fatalf("'ALLOW_TAG_PREFIX' was deprecated, use 'TAG_PREFIX_REGEX' instead")
	}

	return conf, nil
}