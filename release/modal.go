package release

import (
	"context"
	"fmt"
	"os"

	"github.com/google/go-github/github"
)

const (
	SlugRegex            string = `^(?P<owner>[\w,\-,\_\.]+)\/(?P<repo>[\w\,\-\_\.]+)$`
	UnreleasedDefaultTag string = "latest"
)

var (
	UnreleasedRef string = fmt.Sprintf("refs/tags/%v", UnreleasedDefaultTag)
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

type RepositoriesClient interface {
	UploadReleaseAsset(context.Context, string, string, int64, *github.UploadOptions, *os.File) (*github.ReleaseAsset, *github.Response, error)
	CreateRelease(context.Context, string, string, *github.RepositoryRelease) (*github.RepositoryRelease, *github.Response, error)
	DeleteRelease(context.Context, string, string, int64) (*github.Response, error)
	GetReleaseByTag(context.Context, string, string, string) (*github.RepositoryRelease, *github.Response, error)
}

type GitClient interface {
	CreateRef(context.Context, string, string, *github.Reference) (*github.Reference, *github.Response, error)
	DeleteRef(context.Context, string, string, string) (*github.Response, error)
	GetRef(context.Context, string, string, string) (*github.Reference, *github.Response, error)
}
