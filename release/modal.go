package release

import (
	"context"
	"os"

	"github.com/google/go-github/github"
)

const (
	SlugRegex string = `^(?P<owner>[\w,\-,\_\.]+)\/(?P<repo>[\w\,\-\_\.]+)$`
)

type Release struct {
	Name       string
	Slug       *Slug
	Reference  *Reference
	Draft      bool
	PreRelease bool
	Assets     *[]Asset
	Changelog  string
}

type Slug struct {
	Name  string
	Owner string
}

type Reference struct {
	CommitHash string
	Tag        string
	Version    string
}

type Asset struct {
	Name string
	Path string
}

type Client interface {
	UploadReleaseAsset(context.Context, string, string, int64, *github.UploadOptions, *os.File) (*github.ReleaseAsset, *github.Response, error)
	CreateRelease(context.Context, string, string, *github.RepositoryRelease) (*github.RepositoryRelease, *github.Response, error)
}
