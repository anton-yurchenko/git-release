package main

import (
	"git-release/internal/pkg/local"
	"git-release/internal/pkg/remote"
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
	} else {
		log.Warn("environmental variable DRAFT_RELEASE not set, assuming FALSE")
	}

	// get prerelease
	p := os.Getenv("PRE_RELEASE")
	c.Draft = false
	if p == "true" {
		c.Draft = true
	} else {
		log.Warn("environmental variable PRE_RELEASE not set, assuming FALSE")
	}

	// get workspace
	c.Home = os.Getenv("GITHUB_WORKSPACE")
	if c.Home == "" {
		log.Fatal("environmental variable GITHUB_WORKSPACE not defined")
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

	body := "" // TODO: parse CHANGELOG.md

	r.Release.Body = &body

	// prepare releast assets
	// github 'jobs.<job_id>.steps.with.args' does not support arrays, so we need to parse it
	arguments := strings.Split(os.Args[1], "\n") // TODO: parse by both newline and space
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
}
