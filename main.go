package main

import (
	"git-release/release"
	"time"

	"github.com/pkg/errors"
	"github.com/spf13/afero"

	"os"

	log "github.com/sirupsen/logrus"
)

// Version contains current application version
const Version string = "6.0.0"

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
}

// validateEnvironment terminates the execution in case a required environmental variable is not defined.
//
// NOTE: intentionally not a part of init(), otherwise it would terminate the test
// binary of this package before any test had a chance to run.
func validateEnvironment() {
	l := []string{
		"GITHUB_REPOSITORY",
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

	if getToken() == "" {
		log.Fatal("GITHUB_TOKEN is not defined")
	}
}

// getToken returns the GitHub token.
//
// The 'token' input is checked first so that a workflow using
// 'uses: anton-yurchenko/git-release@vN' can rely on the action.yml default and
// omit the env block entirely. A bare 'uses: docker://...' step never reads
// action.yml, so GITHUB_TOKEN remains the way to supply it there.
func getToken() string {
	if v := os.Getenv("INPUT_TOKEN"); v != "" {
		return v
	}

	return os.Getenv("GITHUB_TOKEN")
}

func main() {
	log.Debugf("git-release v%v ", Version)
	validateEnvironment()

	fs := afero.NewOsFs()

	conf, err := GetConfig(fs)
	if err != nil {
		log.Fatal(errors.Wrap(err, "error fetching configuration"))
	}

	rel, err := release.GetRelease(
		fs,
		conf.Assets,
		conf.TagPrefix,
		conf.UnreleasedCreate || conf.UnreleasedDelete,
	)
	if err != nil {
		log.Fatal(errors.Wrap(err, "error fetching release configuration"))
	}

	// NOTE: the title and body are only built for a release that will actually
	// be published. v6 rendered the body before reaching the teardown branch
	// below, so UNRELEASED=delete failed whenever the Unreleased section was
	// empty - it demanded a changelog it was never going to publish.
	if !conf.UnreleasedDelete {
		if rel.Name, err = conf.GetName(rel); err != nil {
			log.Fatal(err)
		}

		if rel.Body, err = conf.GetBody(fs, rel); err != nil {
			log.Fatal(err)
		}
	}

	cli, err := Login(getToken())
	if err != nil {
		log.Fatal(errors.Wrap(err, "login error"))
	}

	if conf.UnreleasedCreate || conf.UnreleasedDelete {
		log.Warnf("deleting precedent release ❗")
		err := rel.DeleteUnreleased(cli.Repositories, cli.Git)
		if err != nil {
			log.Fatal(errors.Wrap(err, "error preparing for Unreleased release update"))
		}

		if conf.UnreleasedDelete {
			return
		}

		if err := rel.UpdateUnreleasedTag(cli.Git); err != nil {
			log.Fatal(errors.Wrapf(err, "error creating %v tag", rel.Reference.Tag))
		}
		time.Sleep(3 * time.Second)
	}

	log.Infof("creating %v release", rel.Name)
	if err := rel.Publish(cli.Repositories); err != nil {
		log.Fatal(err)
	}
}
