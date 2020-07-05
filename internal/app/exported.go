package app

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/anton-yurchenko/git-release/internal/pkg/interfaces"
	"github.com/anton-yurchenko/git-release/internal/pkg/release"
	"github.com/anton-yurchenko/git-release/internal/pkg/repository"
	"github.com/anton-yurchenko/git-release/pkg/changelog"
	"github.com/google/go-github/v32/github"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/afero"
	"golang.org/x/oauth2"
)

// Configuration is a git-release settings struct
type Configuration struct {
	AllowEmptyChangelog bool
	IgnoreChangelog     bool
	AllowTagPrefix      bool
	ReleaseName         string
	ReleaseNamePrefix   string
	ReleaseNamePostfix  string
}

// GetConfig sets validated Release/Changelog configuration and returns github.com Token
func GetConfig(release release.Interface, changes changelog.Interface, fs afero.Fs, args []string) (*Configuration, string, error) {
	conf := new(Configuration)

	token := os.Getenv("GITHUB_TOKEN")
	if token == "" {
		return conf, "", errors.New("'GITHUB_TOKEN' is not defined")
	}

	d := os.Getenv("DRAFT_RELEASE")
	if strings.ToLower(d) == "true" {
		release.EnableDraft()
	} else if strings.ToLower(d) != "false" {
		log.Warn("'DRAFT_RELEASE' is not equal to 'true', assuming 'false'")
	}

	p := os.Getenv("PRE_RELEASE")
	if strings.ToLower(p) == "true" {
		release.EnablePreRelease()
	} else if strings.ToLower(p) != "false" {
		log.Warn("'PRE_RELEASE' is not equal to 'true', assuming 'false'")
	}

	dir := os.Getenv("GITHUB_WORKSPACE")
	if dir == "" {
		log.Fatal("'GITHUB_WORKSPACE' is not defined")
	}

	t := os.Getenv("ALLOW_EMPTY_CHANGELOG")
	if strings.ToLower(t) == "true" {
		log.Warn("'ALLOW_EMPTY_CHANGELOG' enabled")
		conf.AllowEmptyChangelog = true
	}

	t = os.Getenv("ALLOW_TAG_PREFIX")
	if strings.ToLower(t) == "true" {
		log.Warn("'ALLOW_TAG_PREFIX' enabled")
		conf.AllowTagPrefix = true
	}

	rName := os.Getenv("RELEASE_NAME")
	if rName != "" {
		log.Warnf("'RELEASE_NAME' is set to '%v'", rName)
		conf.ReleaseName = rName
	}

	rNamePrefix := os.Getenv("RELEASE_NAME_PREFIX")
	if rNamePrefix != "" {
		log.Warnf("'RELEASE_NAME_PREFIX' is set to '%v'", rNamePrefix)
		conf.ReleaseNamePrefix = rNamePrefix
	}

	rNamePostfix := os.Getenv("RELEASE_NAME_POSTFIX")
	if rNamePostfix != "" {
		log.Warnf("'RELEASE_NAME_POSTFIX' is set to '%v'", rNamePostfix)
		conf.ReleaseNamePostfix = rNamePostfix
	}

	if rName != "" && ((rNamePrefix != "" && rNamePostfix != "") || (rNamePrefix != "" || rNamePostfix != "")) {
		log.Fatal("both 'RELEASE_NAME' and 'RELEASE_NAME_PREFIX'/'RELEASE_NAME_POSTFIX' are set (expected 'RELEASE_NAME' or combination/one of 'RELEASE_NAME_PREFIX' 'RELEASE_NAME_POSTFIX')")
	}

	c := os.Getenv("CHANGELOG_FILE")
	if c == "none" {
		log.Warn("'CHANGELOG_FILE' is set to 'none'")
		conf.IgnoreChangelog = true
	} else if c == "" {
		log.Warn("'CHANGELOG_FILE' is not defined, assuming 'CHANGELOG.md'")
		c = "CHANGELOG.md"
	}

	if !conf.IgnoreChangelog {
		changes.SetFile(fmt.Sprintf("%v/%v", dir, c))
		b, err := IsExists(changes.GetFile(), fs)
		if err != nil {
			log.Fatal(err)
		}

		if !b {
			log.Fatalf("changelog file '%v' not found!", changes.GetFile())
		}
	}

	release.SetAssets(GetAssets(fs, args))

	return conf, token, nil
}

// Login to github.com and return authenticated client
func Login(token string) *github.Client {
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(context.Background(), ts)

	return github.NewClient(tc)
}

// Hydrate fetches project's git repository information
func (c *Configuration) Hydrate(local repository.Interface, version *string, releaseName *string) error {
	if err := local.ReadProjectName(); err != nil {
		return err
	}

	if err := local.ReadCommitHash(); err != nil {
		return err
	}

	if err := local.ReadTag(version, c.AllowTagPrefix); err != nil {
		return err
	}

	if c.ReleaseName != "" {
		*releaseName = c.ReleaseName
	} else if c.ReleaseNamePrefix != "" || c.ReleaseNamePostfix != "" {
		*releaseName = fmt.Sprintf("%v%v%v", c.ReleaseNamePrefix, *local.GetTag(), c.ReleaseNamePostfix)
	} else {
		*releaseName = *local.GetTag()
	}

	return nil
}

// GetReleaseBody populates 'changes.Body' property
// Body will be empty in case version did not match any of the changelog versions.
func (c *Configuration) GetReleaseBody(changes changelog.Interface, fs afero.Fs) error {
	if err := changes.ReadChanges(fs); err != nil {
		return err
	}

	if changes.GetBody() == "" {
		if c.AllowEmptyChangelog {
			log.Warn("changelog does not contain changes for requested project version")
		} else {
			return errors.New("changelog does not contain changes for requested project version")
		}
	}

	return nil
}

// Publish Release on github.com
func (c *Configuration) Publish(repo repository.Interface, release release.Interface, service interfaces.GitHub) error {
	log.Infof("creating release: '%v'", *repo.GetTag())

	errors := make(chan error, len(release.GetAssets()))
	messages := make(chan string, len(release.GetAssets()))

	err := release.Publish(repo, service, messages, errors)
	if err != nil {
		log.Fatal(err)
	}

	for i := 0; i <= (len(release.GetAssets()) - 1); i++ {
		msg := <-messages

		if msg != "" {
			log.Info(msg)
		}
	}

	var failure bool
	for i := 0; i <= (len(release.GetAssets()) - 1); i++ {
		err = <-errors

		if err != nil {
			failure = true
			log.Error(err)
		}
	}

	if failure {
		log.Fatal("error uploading assets (release partially published)")
	}

	return nil
}
