package asset_test

import (
	"fmt"
	"testing"

	"sync"

	"github.com/anton-yurchenko/git-release/internal/pkg/asset"
	"github.com/anton-yurchenko/git-release/internal/pkg/repository"
	"github.com/anton-yurchenko/git-release/mocks"
	"github.com/google/go-github/v32/github"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestSetName(t *testing.T) {
	assert := assert.New(t)
	t.Log("Test Case 1/1 - Functionality")

	m := new(asset.Asset)
	expected := "value"
	m.SetName(expected)

	assert.Equal(expected, m.Name)
}

func TestSetPath(t *testing.T) {
	assert := assert.New(t)
	t.Log("Test Case 1/1 - Functionality")

	m := new(asset.Asset)
	expected := "value"
	m.SetPath(expected)

	assert.Equal(expected, m.Path)
}

func TestUpload(t *testing.T) {
	assert := assert.New(t)
	log.SetLevel(log.FatalLevel)

	var id int64

	// TEST: successful upload
	t.Log("Test Case 1/2 - Functionality")

	m := asset.Asset{
		Name: "file1",
		Path: "asset_test.go",
	}
	id = 1
	repo := new(repository.Repository)
	svc := new(mocks.GitHub)
	wg := new(sync.WaitGroup)
	wg.Add(1)
	msgChan := make(chan string, 1)
	errChan := make(chan error, 1)

	svc.On("UploadReleaseAsset").Return(new(github.ReleaseAsset), new(github.Response), nil).Once()

	m.Upload(id, repo, svc, wg, msgChan, errChan)

	msg := <-msgChan
	err := <-errChan

	assert.Equal(fmt.Sprintf("uploading asset: '%v'", m.Name), msg)
	assert.Equal(nil, err)

	// TEST: failed upload
	t.Log("Test Case 2/2 - Failed Upload")

	m = asset.Asset{
		Name: "file1",
		Path: "asset_test.go",
	}
	id = 2
	repo = new(repository.Repository)
	svc = new(mocks.GitHub)
	wg = new(sync.WaitGroup)
	wg.Add(1)
	msgChan = make(chan string, 1)
	errChan = make(chan error, 1)

	svc.On("UploadReleaseAsset").Return(new(github.ReleaseAsset), new(github.Response), errors.New("value")).Once()

	m.Upload(id, repo, svc, wg, msgChan, errChan)

	msg = <-msgChan
	err = <-errChan

	assert.Equal(fmt.Sprintf("uploading asset: '%v'", m.Name), msg)
	assert.EqualError(err, "value")
}
