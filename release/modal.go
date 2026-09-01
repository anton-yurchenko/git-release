package release

import (
	"context"
	"os"

	"github.com/google/go-github/v78/github"
)

const (
	SlugRegex            string = `^(?P<owner>[\w,\-,\_\.]+)\/(?P<repo>[\w\,\-\_\.]+)$`
	UnreleasedDefaultTag string = "latest"

	// UnreleasedVersion is the sentinel Reference.Version of a rolling release.
	UnreleasedVersion string = "Unreleased"

	// DefaultTagPrefixRegex matches an optional 'v' in front of the version.
	DefaultTagPrefixRegex string = `v?`
)

type Release struct {
	Name       string
	Slug       *Slug
	Reference  *Reference
	Draft      bool
	PreRelease bool
	Assets     *[]Asset
	Body       string
}

type Slug struct {
	Name  string
	Owner string
}

type Reference struct {
	CommitHash string
	Tag        string
	Version    string

	// Semantic version components, parsed out of Version. Empty for a rolling
	// release. Prerelease is the identifier after '-', e.g. "rc.1" - it is what
	// PRE_RELEASE=auto keys off.
	Major      string
	Minor      string
	Patch      string
	Prerelease string
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
	DeleteReleaseAsset(context.Context, string, string, int64) (*github.Response, error)
}

type GitClient interface {
	CreateRef(context.Context, string, string, github.CreateRef) (*github.Reference, *github.Response, error)
	DeleteRef(context.Context, string, string, string) (*github.Response, error)
	GetRef(context.Context, string, string, string) (*github.Reference, *github.Response, error)
}
