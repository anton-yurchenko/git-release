package main

import (
	"context"
	"os"

	"github.com/google/go-github/github"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"golang.org/x/oauth2"
)

// Login to github.com and return authenticated client
func Login(token string) (*github.Client, error) {
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(context.Background(), ts)

	if os.Getenv("GITHUB_API_URL") != "https://api.github.com" && os.Getenv("GITHUB_SERVER_URL") != "https://github.com" {
		log.Info("running on GitHub Enterprise")

		c, err := github.NewEnterpriseClient(os.Getenv("GITHUB_API_URL"), os.Getenv("GITHUB_SERVER_URL"), tc)
		if err != nil {
			return nil, errors.Wrap(err, "error connecting to a private github instance")
		}

		return c, nil
	}

	return github.NewClient(tc), nil
}
