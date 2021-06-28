package main

import (
	"github.com/anton-yurchenko/git-release/release"
	"github.com/pkg/errors"
	"github.com/spf13/afero"

	"os"

	log "github.com/sirupsen/logrus"
)

// Version contains current application version
const Version string = "4.1.0"

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
		conf.ReleaseNameSuffix,
		conf.Unreleased,
	)
	if err != nil {
		log.Fatal(errors.Wrap(err, "error fetching release configuration"))
	}

	if conf.ChangelogFile != "" {
		rel.Changelog, err = conf.GetChangelog(fs, rel)
		if err != nil {
			log.Fatal(errors.Wrap(err, "error reading changelog"))
		}
	}

	cli, err := Login(os.Getenv("GITHUB_TOKEN"))
	if err != nil {
		log.Fatal(errors.Wrap(err, "login error"))
	}

	if conf.Unreleased {
		if err := rel.DeleteUnreleased(cli.Repositories, cli.Git); err != nil {
			log.Fatal(errors.Wrap(err, "error preparing for Unreleased release update"))
		}
	}

	log.Infof("creating release %v", rel.Name)
	if err := rel.Publish(cli.Repositories); err != nil {
		log.Fatal(err)
	}
}
