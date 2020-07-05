package asset

import (
	"context"
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/anton-yurchenko/git-release/internal/pkg/interfaces"
	"github.com/anton-yurchenko/git-release/internal/pkg/repository"
	"github.com/google/go-github/v32/github"
)

// Asset represents a single release artifact
type Asset struct {
	Name string
	Path string
}

// Interface of an 'Asset'
type Interface interface {
	Upload(int64, repository.Interface, interfaces.GitHub, *sync.WaitGroup, chan string, chan error)
	SetName(string)
	SetPath(string)
}

// Upload an asset to pre-created github.com Release
func (a *Asset) Upload(id int64, repo repository.Interface, service interfaces.GitHub, wg *sync.WaitGroup, messages chan string, results chan error) {
	defer wg.Done()
	messages <- fmt.Sprintf("uploading asset: '%v'", a.Name)

	content, err := os.Open(a.Path)
	if err != nil {
		results <- err
		return
	}

	_, _, err = service.UploadReleaseAsset(
		context.Background(),
		repo.GetOwner(),
		repo.GetProject(),
		id,
		&github.UploadOptions{
			Name: strings.ReplaceAll(a.Name, "/", "-"),
		},
		content,
	)
	if err != nil {
		results <- err
		return
	}

	results <- nil
}

// SetName sets provided string as an asset Name
func (a *Asset) SetName(name string) {
	a.Name = name
}

// SetPath sets provided string as an asset Path
func (a *Asset) SetPath(path string) {
	a.Path = path
}
