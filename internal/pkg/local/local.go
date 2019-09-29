package local

import (
	"errors"
	"git-release/internal/pkg/remote"
	"os"
	"regexp"
	"strings"
)

// GetDetails returns local git information
func GetDetails(r *remote.Remote) error {
	repoName, err := getRepoName()
	if err != nil {
		return err
	}

	hash, err := getCommitHash()
	if err != nil {
		return err
	}

	tag, err := getVersionTag()
	if err != nil {
		return err
	}

	r.Owner = repoName["owner"]
	r.Repository = repoName["name"]
	r.Release.CommitHash = &hash
	r.Release.Tag = &tag
	r.Release.Name = &tag

	return nil
}

func getRepoName() (map[string]string, error) {
	o := os.Getenv("GITHUB_REPOSITORY")
	if o == "" {
		return map[string]string{}, errors.New("environmental variable GITHUB_REPOSITORY not defined")
	}

	repo := make(map[string]string)

	repo["owner"] = strings.Split(o, "/")[0]
	repo["name"] = strings.Split(o, "/")[1]

	return repo, nil
}

func getCommitHash() (string, error) {
	o := os.Getenv("GITHUB_SHA")
	if o == "" {
		return "", errors.New("environmental variable GITHUB_SHA not defined")
	}

	return o, nil
}

func getVersionTag() (string, error) {
	o := os.Getenv("GITHUB_REF")
	if o == "" {
		return "", errors.New("environmental variable GITHUB_REF not defined")
	}

	regex := regexp.MustCompile("refs/tags/v?[0-9]+.[0-9]+.[0-9]+")
	if regex.MatchString(o) {
		return strings.Split(o, "/")[2], nil
	}

	return "", errors.New("no matching tags found. expected to match regex 'v?[0-9]+.[0-9]+.[0-9]+'")
}
