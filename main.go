package main

import (
	"github.com/anton-yurchenko/git-release/internal/app"
	"github.com/anton-yurchenko/git-release/internal/pkg/release"
	"github.com/anton-yurchenko/git-release/internal/pkg/repository"
	"github.com/anton-yurchenko/git-release/pkg/changelog"
	"github.com/spf13/afero"

	"os"

	log "github.com/sirupsen/logrus"
)

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

	log.Info("version: ", app.Version)
}

func main() {
	fs := afero.NewOsFs()
	repo := new(repository.Repository)
	release := new(release.Release)
	release.Changes = new(changelog.Changes)

	conf, err := app.GetConfig(release, release.Changes, fs, os.Args[1:])
	if err != nil {
		log.Fatal(err)
	}

	cli, err := app.Login(os.Getenv("GITHUB_TOKEN"))
	if err != nil {
		log.Fatal(err)
	}

	if err := conf.Hydrate(repo, &release.Changes.Version, &release.Name); err != nil {
		log.Fatal(err)
	}

	if !conf.IgnoreChangelog {
		if err = conf.GetReleaseBody(release.Changes, fs); err != nil {
			log.Fatal(err)
		}
	}

	if err = conf.Publish(repo, release, cli.Repositories); err != nil {
		log.Fatal(err)
	}
}
