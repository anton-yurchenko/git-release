package main

import (
	"io"
	"testing"

	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestLogin(t *testing.T) {
	a := assert.New(t)
	log.SetOutput(io.Discard)

	t.Run("github.com", func(t *testing.T) {
		t.Setenv("GITHUB_API_URL", "https://api.github.com")
		t.Setenv("GITHUB_SERVER_URL", "https://github.com")

		c, err := Login("token")
		a.NoError(err)
		a.Equal("https://api.github.com/", c.BaseURL.String())
		a.Equal("https://uploads.github.com/", c.UploadURL.String())
	})

	// On GitHub Enterprise the API lives under /api/v3 and uploads under
	// /api/uploads. v6 only appended a trailing slash to whatever it was given,
	// so asset uploads were POSTed to the bare host and failed.
	t.Run("github enterprise", func(t *testing.T) {
		t.Setenv("GITHUB_API_URL", "https://github.example.com/api/v3")
		t.Setenv("GITHUB_SERVER_URL", "https://github.example.com")

		c, err := Login("token")
		a.NoError(err)
		a.Equal("https://github.example.com/api/v3/", c.BaseURL.String())
		a.Equal("https://github.example.com/api/uploads/", c.UploadURL.String())
	})

	t.Run("malformed enterprise url", func(t *testing.T) {
		t.Setenv("GITHUB_API_URL", "://not a url")
		t.Setenv("GITHUB_SERVER_URL", "://not a url")

		_, err := Login("token")
		a.ErrorContains(err, "error connecting to a private github instance")
	})
}
