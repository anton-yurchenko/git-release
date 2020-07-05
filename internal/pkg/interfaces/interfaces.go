package interfaces

import (
	"context"
	"os"

	"github.com/google/go-github/v32/github"
)

// GitHub is a 'github.client' interface
type GitHub interface {
	UploadReleaseAsset(context.Context, string, string, int64, *github.UploadOptions, *os.File) (*github.ReleaseAsset, *github.Response, error)
	CreateRelease(context.Context, string, string, *github.RepositoryRelease) (*github.RepositoryRelease, *github.Response, error)
}
