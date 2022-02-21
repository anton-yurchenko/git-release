package main

import (
	"strings"

	"git-release/release"

	"github.com/pkg/errors"
	"github.com/spf13/afero"

	"os"

	log "github.com/sirupsen/logrus"
)

// Version contains current application version
const Version string = "4.2.4"

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
		conf.UnreleasedCreate || conf.UnreleasedDelete,
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

	if conf.UnreleasedCreate || conf.UnreleasedDelete {
		err := rel.DeleteUnreleased(cli.Repositories, cli.Git)
		if err != nil {
			if !strings.Contains(err.Error(), "precedent release not found") {
				log.Fatal(errors.Wrap(err, "error preparing for Unreleased release update"))
			}

			log.Warn(err.Error())
		} else {
			log.Warnf("precedent release deleted ❗")
		}

		if conf.UnreleasedDelete {
			return
		}

		if err := rel.UpdateUnreleasedTag(cli.Git); err != nil {
			log.Fatal(errors.Wrapf(err, "error creating %v tag", rel.Reference.Tag))
		}
	}

	log.Infof("creating %v release", rel.Name)
	if err := rel.Publish(cli.Repositories); err != nil {
		log.Fatal(err)
	}
}
