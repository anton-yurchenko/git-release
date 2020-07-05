package mocks

import (
	context "context"

	github "github.com/google/go-github/v32/github"

	mock "github.com/stretchr/testify/mock"

	os "os"
)

// GitHub is an mock type for the GitHub type
type GitHub struct {
	mock.Mock
}

// CreateRelease provides a stub function
func (_m *GitHub) CreateRelease(context.Context, string, string, *github.RepositoryRelease) (*github.RepositoryRelease, *github.Response, error) {
	args := _m.Called()

	return args.Get(0).(*github.RepositoryRelease), args.Get(1).(*github.Response), args.Error(2)
}

// UploadReleaseAsset provides a stub function
func (_m *GitHub) UploadReleaseAsset(context.Context, string, string, int64, *github.UploadOptions, *os.File) (*github.ReleaseAsset, *github.Response, error) {
	args := _m.Called()

	return args.Get(0).(*github.ReleaseAsset), args.Get(1).(*github.Response), args.Error(2)
}
