package release

import (
	"context"
	"fmt"
	"os"
	"regexp"
	"strings"
	"sync"
	"time"

	changelog "github.com/anton-yurchenko/go-changelog"
	"github.com/google/go-github/github"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/afero"
)

func GetRelease(fs afero.Fs, args []string, tagPrefix, name, namePrefix, nameSuffix string, unreleased bool) (*Release, error) {
	release := new(Release)

	if strings.ToLower(os.Getenv("DRAFT_RELEASE")) == "true" {
		release.Draft = true
	}

	if strings.ToLower(os.Getenv("PRE_RELEASE")) == "true" || unreleased {
		release.PreRelease = true
	}

	var err error
	release.Assets, err = GetAssets(fs, args)
	if err != nil {
		return nil, errors.Wrap(err, "error retrieving release assets")
	}

	release.Reference, err = GetReference(tagPrefix, unreleased)
	if err != nil {
		return nil, errors.Wrap(err, "error retrieving source code reference (control tag prefix via env.var TAG_PREFIX_REGEX)")
	}

	release.Slug, err = GetSlug()
	if err != nil {
		return nil, errors.Wrap(err, "error retrieving repository slug")
	}

	if name != "" {
		release.Name = name
	} else if unreleased {
		release.Name = "Latest"
	} else {
		release.Name = fmt.Sprintf("%v%v%v", namePrefix, release.Reference.Tag, nameSuffix)
	}

	return release, nil
}

// GetReference loads a codebase references from workspace
func GetReference(prefix string, unreleased bool) (*Reference, error) {
	if os.Getenv("GITHUB_REF") == "" {
		return nil, errors.New("GITHUB_REF is not defined")
	} else if os.Getenv("GITHUB_REF") == UnreleasedRef {
		return nil, errors.New("workflow configuration error detected: trigger loop (triggering tag will be recreated and trigger the workflow again)")
	}

	if os.Getenv("GITHUB_SHA") == "" {
		return nil, errors.New("GITHUB_SHA is not defined")
	}

	if unreleased {
		tag := UnreleasedDefaultTag

		if os.Getenv("UNRELEASED_TAG") != "" {
			tag = os.Getenv("UNRELEASED_TAG")
		}

		return &Reference{
			CommitHash: os.Getenv("GITHUB_SHA"),
			Tag:        tag,
			Version:    "Unreleased",
		}, nil
	}

	var expression string
	if prefix != "" {
		expression = fmt.Sprintf("^refs/tags/(?P<prefix>%v)%v$", prefix, changelog.SemVerRegex)
	} else {
		expression = fmt.Sprintf("^refs/tags/[v]?%v$", changelog.SemVerRegex)
	}
	regex := regexp.MustCompile(expression)

	if regex.MatchString(os.Getenv("GITHUB_REF")) {
		var version string
		if prefix != "" {
			versionRegex := regexp.MustCompile(fmt.Sprintf("^refs/tags/(?P<prefix>%v)(?P<version>.*)$", prefix))
			if versionRegex.MatchString(os.Getenv("GITHUB_REF")) {
				version = versionRegex.ReplaceAllString(os.Getenv("GITHUB_REF"), "${2}")
			} else {
				version = strings.TrimPrefix(os.Getenv("GITHUB_REF"), "refs/tags/")
			}
		} else {
			version = strings.TrimPrefix(strings.TrimPrefix(os.Getenv("GITHUB_REF"), "refs/tags/"), "v")
		}

		return &Reference{
			CommitHash: os.Getenv("GITHUB_SHA"),
			Tag:        strings.Join(strings.Split(os.Getenv("GITHUB_REF"), "/")[2:], "/"),
			Version:    version,
		}, nil
	}

	return nil, errors.New(fmt.Sprintf("malformed env.var GITHUB_REF: expected to match regex '%v', got '%v'", expression, os.Getenv("GITHUB_REF")))
}

// GetSlug loads project information from a workspace
func GetSlug() (*Slug, error) {
	if os.Getenv("GITHUB_REPOSITORY") == "" {
		return nil, errors.New("GITHUB_REPOSITORY is not defined")
	}

	i := os.Getenv("GITHUB_REPOSITORY")
	regex := regexp.MustCompile(SlugRegex)

	if regex.MatchString(i) {
		return &Slug{
			Owner: strings.Split(i, "/")[0],
			Name:  strings.Split(i, "/")[1],
		}, nil
	}

	return nil, errors.New(fmt.Sprintf("malformed GITHUB_REPOSITORY (expected '%v', received '%v')", SlugRegex, i))
}

// Publish will create a GitHub release and upload assets to it
func (r *Release) Publish(cli RepositoriesClient) error {
	// create release
	o, _, err := cli.CreateRelease(
		context.Background(),
		r.Slug.Owner,
		r.Slug.Name,
		&github.RepositoryRelease{
			Name:            &r.Name,
			TagName:         &r.Reference.Tag,
			TargetCommitish: &r.Reference.CommitHash,
			Body:            &r.Changelog,
			Draft:           &r.Draft,
			Prerelease:      &r.PreRelease,
		},
	)
	if err != nil {
		return err
	}

	log.Info("release created successfully 🎉")

	// upload assets
	if r.Assets != nil {
		errs := make(chan error, len(*r.Assets))
		messages := make(chan string, len(*r.Assets))

		wg := new(sync.WaitGroup)
		wg.Add(len(*r.Assets))

		for _, a := range *r.Assets {
			asset := a
			go asset.Upload(r, cli, o.GetID(), messages, errs, wg)
		}

		for i := 0; i <= (len(*r.Assets) - 1); i++ {
			msg := <-messages

			if msg != "" {
				log.Info(msg)
			}
		}

		var failure bool
		for i := 0; i <= (len(*r.Assets) - 1); i++ {
			err = <-errs

			if err != nil {
				failure = true
				log.Error(err)
			}
		}

		wg.Wait()

		if failure {
			return errors.New("error uploading assets")
		}

		log.Info("assets uploaded successfully 🎉")
	}

	return nil
}

// DeleteUnreleased prepares a repository for an update of an existing Unreleased release.
// This includes a deletion of previous release and recreation of the tag.
func (r *Release) DeleteUnreleased(repoCli RepositoriesClient, gitCli GitClient) error {
	tag := fmt.Sprintf("refs/tags/%v", r.Reference.Tag)

	previous, _, err := repoCli.GetReleaseByTag(
		context.Background(),
		r.Slug.Owner,
		r.Slug.Name,
		r.Reference.Tag,
	)
	if err != nil {
		if !strings.Contains(err.Error(), "404 Not Found") {
			return errors.Wrapf(err, "error retrieving a precedent release with a tag %v", r.Reference.Tag)
		}

		return errors.New("precedent release not found")
	}

	_, err = repoCli.DeleteRelease(
		context.Background(),
		r.Slug.Owner,
		r.Slug.Name,
		previous.GetID(),
	)
	if err != nil {
		return errors.Wrap(err, "error deleting precedent release")
	}

	_, err = gitCli.DeleteRef(
		context.Background(),
		r.Slug.Owner,
		r.Slug.Name,
		tag,
	)
	if err != nil {
		return errors.Wrap(err, "error deleting precedent tag")
	}

	for i := 0; i < 3; i++ {
		_, _, err := gitCli.GetRef(
			context.Background(),
			r.Slug.Owner,
			r.Slug.Name,
			tag,
		)
		if err != nil {
			if strings.Contains(err.Error(), "404 Not Found") {
				break
			}

			return errors.Wrap(err, "error fetching precedent tag")
		}

		time.Sleep(3 * time.Second)
	}

	return nil
}

func (r *Release) UpdateUnreleasedTag(gitCli GitClient) error {
	tag := fmt.Sprintf("refs/tags/%v", r.Reference.Tag)

	_, _, err := gitCli.CreateRef(
		context.Background(),
		r.Slug.Owner,
		r.Slug.Name,
		&github.Reference{
			Ref: &tag,
			Object: &github.GitObject{
				SHA: &r.Reference.CommitHash,
			},
		},
	)

	return err
}
