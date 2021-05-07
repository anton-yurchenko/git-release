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
	"github.com/google/go-github/v35/github"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/afero"
	"golang.org/x/oauth2"
)

const ()

// Configuration is a git-release settings struct
type Configuration struct {
	AllowEmptyChangelog bool
	IgnoreChangelog     bool
	AllowTagPrefix      bool
	ReleaseName         string
	ReleaseNamePrefix   string
	ReleaseNameSuffix   string
}

// GetConfig sets validated Release/Changelog configuration and returns github.com Token
func GetConfig(release release.Interface, changes changelog.Interface, fs afero.Fs, args []string) (*Configuration, error) {
	conf := new(Configuration)

	l := []string{
		"GITHUB_TOKEN",
		"GITHUB_WORKSPACE",
		"GITHUB_API_URL",
		"GITHUB_SERVER_URL",
		"GITHUB_REF",
		"GITHUB_SHA",
	}

	for _, v := range l {
		if os.Getenv(v) == "" {
			return conf, errors.New(fmt.Sprintf("'%v' is not defined", v))
		}
	}

	if strings.ToLower(os.Getenv("DRAFT_RELEASE")) == "true" {
		release.EnableDraft()
	}

	if strings.ToLower(os.Getenv("PRE_RELEASE")) == "true" {
		release.EnablePreRelease()
	}

	if strings.ToLower(os.Getenv("ALLOW_EMPTY_CHANGELOG")) == "true" {
		conf.AllowEmptyChangelog = true
	}

	if strings.ToLower(os.Getenv("ALLOW_TAG_PREFIX")) == "true" {
		conf.AllowTagPrefix = true
	}

	if os.Getenv("RELEASE_NAME") != "" {
		conf.ReleaseName = os.Getenv("RELEASE_NAME")
	}

	if os.Getenv("RELEASE_NAME_PREFIX") != "" {
		conf.ReleaseNamePrefix = os.Getenv("RELEASE_NAME_PREFIX")
	}

	if os.Getenv("RELEASE_NAME_SUFFIX") != "" {
		conf.ReleaseNameSuffix = os.Getenv("RELEASE_NAME_SUFFIX")
	}

	if conf.ReleaseName != "" && ((conf.ReleaseNamePrefix != "" && conf.ReleaseNameSuffix != "") || (conf.ReleaseNamePrefix != "" || conf.ReleaseNameSuffix != "")) {
		log.Fatal("both 'RELEASE_NAME' and 'RELEASE_NAME_PREFIX'/'RELEASE_NAME_SUFFIX' are set (expected 'RELEASE_NAME' or combination/one of 'RELEASE_NAME_PREFIX' 'RELEASE_NAME_SUFFIX')")
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
		changes.SetFile(fmt.Sprintf("%v/%v", os.Getenv("GITHUB_WORKSPACE"), c))
		b, err := IsExists(changes.GetFile(), fs)
		if err != nil {
			log.Fatal(err)
		}

		if !b {
			log.Fatalf("changelog file '%v' not found!", changes.GetFile())
		}
	}

	release.SetAssets(GetAssets(fs, args))

	return conf, nil
}

// Login to github.com and return authenticated client
func Login(token string) (*github.Client, error) {
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(context.Background(), ts)

	if os.Getenv("GITHUB_API_URL") != "https://api.github.com" && os.Getenv("GITHUB_SERVER_URL") != "https://github.com" {
		log.Info("running on GitHub Enterprise")

		if os.Getenv("GODEBUG") != "" {
			log.Debug("baseURL: %v, uploadURL: %v", os.Getenv("GITHUB_API_URL"), os.Getenv("GITHUB_SERVER_URL"))
		}

		c, err := github.NewEnterpriseClient(os.Getenv("GITHUB_API_URL"), os.Getenv("GITHUB_SERVER_URL"), tc)
		if err != nil {
			return nil, errors.Wrap(err, "error connecting to a private github instance")
		}

		return c, nil
	}

	return github.NewClient(tc), nil
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
	} else if c.ReleaseNamePrefix != "" || c.ReleaseNameSuffix != "" {
		*releaseName = fmt.Sprintf("%v%v%v", c.ReleaseNamePrefix, *local.GetTag(), c.ReleaseNameSuffix)
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
