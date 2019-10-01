package main

import (
	"git-release/internal/pkg/local"
	"git-release/internal/pkg/remote"
	"git-release/pkg/changelog"
	"strings"

	"os"
	"path/filepath"

	log "github.com/sirupsen/logrus"
)

type config struct {
	Token      string
	Draft      bool
	PreRelease bool
	Home       string
	Changelog  string
}

func init() {
	// set logger
	log.SetReportCaller(false)
	log.SetFormatter(&log.TextFormatter{
		ForceColors:            true,
		FullTimestamp:          true,
		DisableLevelTruncation: true,
		DisableTimestamp:       true,
	})
	log.SetOutput(os.Stdout)
	log.SetLevel(log.InfoLevel)
}

func getConfig() config {
	// get token
	var c config
	c.Token = os.Getenv("GITHUB_TOKEN")
	if c.Token == "" {
		log.Fatal("environmental variable GITHUB_TOKEN not defined")
	}

	// get draft
	d := os.Getenv("DRAFT_RELEASE")
	c.Draft = false
	if d == "true" {
		c.Draft = true
	} else if d != "false" {
		log.Warn("environmental variable DRAFT_RELEASE not set, assuming FALSE")
	}

	// get prerelease
	p := os.Getenv("PRE_RELEASE")
	c.Draft = false
	if p == "true" {
		c.Draft = true
	} else if p != "false" {
		log.Warn("environmental variable PRE_RELEASE not set, assuming FALSE")
	}

	// get workspace
	c.Home = os.Getenv("GITHUB_WORKSPACE")
	if c.Home == "" {
		log.Fatal("environmental variable GITHUB_WORKSPACE not defined")
	}

	// get changelog
	c.Changelog = os.Getenv("CHANGELOG_FILE")
	if c.Changelog == "" {
		log.Warn("environmental variable CHANGELOG_FILE not set, assuming 'CHANGELOG.md'")
		c.Changelog = "CHANGELOG.md"
	}

	return c
}

func main() {
	conf := getConfig()

	// authenticate
	r := remote.Authenticate(conf.Token)

	// get details
	r.Release.Draft = &conf.Draft
	r.Release.PreRelease = &conf.PreRelease

	err := local.GetDetails(&r)
	if err != nil {
		log.Fatal(err)
	}

	// get changelog
	log.Infof("reading changelog: %+s", conf.Changelog)
	r.Release.Body, err = changelog.GetBody(*r.Release.Name, conf.Changelog)
	if *r.Release.Body == "" {
		log.Warn("creating release with empty body")
	}
	if err != nil {
		log.Warn(err)
	}

	// prepare releast assets
	// github 'jobs.<job_id>.steps.with.args' does not support arrays, so we need to parse it
	arguments := strings.Split(os.Args[1], "\n")
	for _, argument := range arguments {
		r.Assets = append(r.Assets, remote.Asset{
			Name: filepath.Base(argument),
			Path: conf.Home + "/" + argument,
		})
	}

	err = r.Publish()

	if err != nil {
		log.Fatal(err)
	}

	log.Infof("release '%+s' published", *r.Release.Name)
}
