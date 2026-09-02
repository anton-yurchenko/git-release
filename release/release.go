package release

import (
	"context"
	"fmt"
	"os"
	"regexp"
	"strings"
	"sync"
	"time"

	"git-release/env"

	changelog "github.com/anton-yurchenko/go-changelog"
	"github.com/google/go-github/v78/github"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/afero"
)

func GetRelease(fs afero.Fs, assets []string, tagPrefix string, unreleased bool) (*Release, error) {
	release := new(Release)

	var err error
	if release.Draft, err = env.Bool("DRAFT_RELEASE"); err != nil {
		return nil, err
	}

	release.Assets, err = GetAssets(fs, assets)
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

	// A rolling release is always a pre-release.
	preRelease, err := env.Bool("PRE_RELEASE")
	if err != nil {
		return nil, err
	}

	release.PreRelease = preRelease || unreleased

	// The default title. NAME_TEMPLATE, when set, replaces it.
	release.Name = release.Reference.Tag
	if unreleased {
		release.Name = "Latest"
	}

	return release, nil
}

// GetReference loads a codebase references from workspace
func GetReference(prefix string, unreleased bool) (*Reference, error) {
	ref := os.Getenv("GITHUB_REF")
	if ref == "" {
		return nil, errors.New("GITHUB_REF is not defined")
	}

	if os.Getenv("GITHUB_SHA") == "" {
		return nil, errors.New("GITHUB_SHA is not defined")
	}

	if unreleased {
		tag := UnreleasedDefaultTag
		if v := env.Get("UNRELEASED_TAG"); v != "" {
			tag = v
		}

		// The rolling tag is deleted and recreated on every run, so a workflow
		// triggered by that tag would trigger itself forever.
		//
		// NOTE: the guard is scoped to the resolved tag, and only in unreleased
		// mode. v6 compared against a hardcoded 'refs/tags/latest', which both
		// missed a custom UNRELEASED_TAG and refused an ordinary release of a
		// tag that happened to be named 'latest'.
		if ref == fmt.Sprintf("refs/tags/%v", tag) {
			return nil, errors.New("workflow configuration error detected: trigger loop (triggering tag will be recreated and trigger the workflow again)")
		}

		return &Reference{
			CommitHash: os.Getenv("GITHUB_SHA"),
			Tag:        tag,
			Version:    UnreleasedVersion,
		}, nil
	}

	if prefix == "" {
		prefix = DefaultTagPrefixRegex
	}

	// NOTE: the prefix is wrapped non-capturing and the version is read by NAME.
	// v6 built '(?P<prefix>%v)' and then extracted the version with a positional
	// '${2}', so a capture group in a user's own regex shifted the numbering:
	// TAG_PREFIX_REGEX='(rel|pre)-' on 'refs/tags/rel-1.2.3' silently yielded
	// the version 'rel'.
	expression := fmt.Sprintf(`^refs/tags/(?:%v)(?P<version>%v)$`, prefix, changelog.SemVerRegex)

	// NOTE: Compile, not MustCompile - the pattern is user input, and a
	// configuration mistake must not panic.
	regex, err := regexp.Compile(expression)
	if err != nil {
		return nil, errors.Wrap(err, "malformed env.var TAG_PREFIX_REGEX")
	}

	match := regex.FindStringSubmatch(ref)
	if match == nil {
		return nil, errors.Errorf("malformed env.var GITHUB_REF: expected to match regex '%v', got '%v'", expression, ref)
	}

	version := match[regex.SubexpIndex("version")]

	// Re-match the extracted version on its own to split it into components.
	// Doing it separately keeps the indexes independent of whatever groups the
	// user's prefix introduced.
	semver := regexp.MustCompile(`^` + changelog.SemVerRegex + `$`)
	parts := semver.FindStringSubmatch(version)

	return &Reference{
		CommitHash: os.Getenv("GITHUB_SHA"),
		Tag:        strings.Join(strings.Split(ref, "/")[2:], "/"),
		Version:    version,
		Major:      parts[1],
		Minor:      parts[2],
		Patch:      parts[3],
		Prerelease: parts[4],
	}, nil
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
			Body:            &r.Body,
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

		wg := new(sync.WaitGroup)
		wg.Add(len(*r.Assets))

		for _, a := range *r.Assets {
			asset := a
			go asset.Upload(r, cli, o.GetID(), errs, wg)
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

	if err == nil {
		_, err = repoCli.DeleteRelease(
			context.Background(),
			r.Slug.Owner,
			r.Slug.Name,
			previous.GetID(),
		)
		if err != nil {
			return errors.Wrap(err, "error deleting precedent release")
		}
	} else if !strings.Contains(err.Error(), "404 Not Found") {
		return errors.Wrapf(err, "error retrieving a precedent release with a tag %v", r.Reference.Tag)
	} else {
		log.Warn("precedent release not found")
	}

	_, err = gitCli.DeleteRef(
		context.Background(),
		r.Slug.Owner,
		r.Slug.Name,
		tag,
	)
	if err == nil {
		// tag deletion takes some time to be reflected
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
	} else if !strings.Contains(err.Error(), "422 Reference does not exist") {
		return errors.Wrap(err, "error deleting precedent tag")
	} else {
		log.Warn("precedent tag not found")
	}

	return nil
}

func (r *Release) UpdateUnreleasedTag(gitCli GitClient) error {
	tag := fmt.Sprintf("refs/tags/%v", r.Reference.Tag)

	_, _, err := gitCli.CreateRef(
		context.Background(),
		r.Slug.Owner,
		r.Slug.Name,
		github.CreateRef{
			Ref: tag,
			SHA: r.Reference.CommitHash,
		},
	)

	return err
}
