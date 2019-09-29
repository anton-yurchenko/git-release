package remote

import (
	"context"

	"os"
	"sync"

	"errors"

	"github.com/google/go-github/github"
	log "github.com/sirupsen/logrus"
	"golang.org/x/oauth2"
)

// Remote git repository and release configuration
type Remote struct {
	Client     *github.Client
	Owner      string
	Repository string
	Release    Release
	Assets     []Asset
}

// Release information
type Release struct {
	ID         int64
	Tag        *string
	CommitHash *string
	Name       *string
	Body       *string
	Draft      *bool
	PreRelease *bool
}

// Asset for an upload
type Asset struct {
	Name string
	Path string
}

// Authenticate against github.com with access token
func Authenticate(token string) Remote {
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(context.Background(), ts)

	return Remote{Client: github.NewClient(tc)}
}

// Publish release and upload assets
func (r *Remote) Publish() error {
	log.Info("creating release: ", *r.Release.Name)

	// create release
	rel, _, err := r.Client.Repositories.CreateRelease(context.Background(), r.Owner, r.Repository, &github.RepositoryRelease{
		TagName:         r.Release.Tag,
		TargetCommitish: r.Release.CommitHash,
		Name:            r.Release.Name,
		Body:            r.Release.Body,
		Draft:           r.Release.Draft,
		Prerelease:      r.Release.PreRelease,
	})

	if err != nil {
		return err
	}

	r.Release.ID = rel.GetID()

	// upload assets
	results := make(chan error, len(r.Assets))
	var wg sync.WaitGroup
	wg.Add(len(r.Assets))
	for _, asset := range r.Assets {
		go r.upload(asset, &wg, results)
	}
	wg.Wait()

	failed := false
	for i := 0; i <= len(r.Assets)-1; i++ {
		result := <-results
		if result != nil {
			log.Errorf("error uploading file %s: %s", r.Assets[i].Name, result)
			failed = true
		}
	}

	if failed {
		return errors.New("some assets were not uploaded, please validate github release manually")
	}

	return nil
}

func (r *Remote) upload(file Asset, wg *sync.WaitGroup, result chan error) {
	defer wg.Done()
	log.Info("uploading asset: ", file.Name)

	content, err := os.Open(file.Path)

	if err != nil {
		result <- err
		return
	}

	_, _, err = r.Client.Repositories.UploadReleaseAsset(context.Background(), r.Owner, r.Repository, r.Release.ID, &github.UploadOptions{
		Name: file.Name,
	}, content)

	if err != nil {
		result <- err
		return
	}

	result <- nil
}
