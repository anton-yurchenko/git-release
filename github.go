package main

import (
	"os"

	"github.com/google/go-github/v78/github"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

// Login to github.com and return authenticated client
func Login(token string) (*github.Client, error) {
	cli := github.NewClient(nil).WithAuthToken(token)

	apiURL := os.Getenv("GITHUB_API_URL")
	serverURL := os.Getenv("GITHUB_SERVER_URL")

	if apiURL != "https://api.github.com" && serverURL != "https://github.com" {
		log.Info("running on GitHub Enterprise")

		c, err := cli.WithEnterpriseURLs(apiURL, serverURL)
		if err != nil {
			return nil, errors.Wrap(err, "error connecting to a private github instance")
		}

		return c, nil
	}

	return cli, nil
}
