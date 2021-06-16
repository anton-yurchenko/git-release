package main

import (
	"fmt"

	"github.com/anton-yurchenko/git-release/release"
	"github.com/anton-yurchenko/go-changelog"
	"github.com/pkg/errors"
	"github.com/spf13/afero"

	"os"

	log "github.com/sirupsen/logrus"
)

// Version contains current application version
const Version string = "4.0.0"

func init() {
	log.SetReportCaller(false)
	log.SetFormatter(&log.TextFormatter{
		ForceColors:            true,
		FullTimestamp:          true,
		DisableLevelTruncation: true,
		DisableTimestamp:       true,
	})
	log.SetOutput(os.Stdout)
	log.SetLevel(log.DebugLevel)

	log.Debugf("git-release v%v ", Version)

	l := []string{
		"GITHUB_REPOSITORY",
		"GITHUB_TOKEN",
		"GITHUB_WORKSPACE",
		"GITHUB_API_URL",
		"GITHUB_SERVER_URL",
		"GITHUB_REF",
		"GITHUB_SHA",
	}

	for _, v := range l {
		if os.Getenv(v) == "" {
			log.Fatalf("%v is not defined", v)
		}
	}
}

func main() {
	fs := afero.NewOsFs()

	conf, err := GetConfig(fs)
	if err != nil {
		log.Fatal(errors.Wrap(err, "error fetching configuration"))
	}

	rel, err := release.GetRelease(
		fs,
		os.Args[1:],
		conf.TagPrefix,
		conf.ReleaseName,
		conf.ReleaseNamePrefix,
		conf.ReleaseNameSuffix)
	if err != nil {
		log.Fatal(errors.Wrap(err, "error fetching release configuration"))
	}

	if conf.ChangelogFile != "" {
		p, err := changelog.NewParserWithFilesystem(fs, conf.ChangelogFile)
		if err != nil {
			log.Fatal(errors.Wrap(err, "error loading changelog file"))
		}

		c, err := p.Parse()
		if err != nil {
			log.Fatal(errors.Wrap(err, "error parsing changelog file"))
		}

		r := c.GetRelease(rel.Reference.Version)
		if r == nil {
			msg := fmt.Sprintf("changelog file does not contain version %v", rel.Reference.Version)

			if !conf.AllowEmptyChangelog {
				log.Fatal(msg)
			} else {
				log.Warn(msg)
			}
		} else {
			rel.Changelog = r.Changes.ToString()
		}
	}

	cli, err := Login(os.Getenv("GITHUB_TOKEN"))
	if err != nil {
		log.Fatal(errors.Wrap(err, "login error"))
	}

	log.Infof("creating release %v", rel.Name)
	if err := rel.Publish(cli.Repositories); err != nil {
		log.Fatal(err)
	}
}
