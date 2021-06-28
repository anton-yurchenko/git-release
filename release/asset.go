package release

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/google/go-github/github"
	"github.com/spf13/afero"
)

// GetAssets returns validated assets supplied via 'args'
func GetAssets(fs afero.Fs, args []string) (*[]Asset, error) {
	assets := make([]Asset, 0)
	arguments := make([]string, 0)

	for _, arg := range args {
		if len(strings.Split(arg, " ")) > 1 {
			arguments = append(arguments, strings.Split(arg, " ")...)
		} else if len(strings.Split(arg, "\n")) > 1 {
			arguments = append(arguments, strings.Split(arg, "\n")...)
		} else if len(strings.Split(arg, ",")) > 1 {
			arguments = append(arguments, strings.Split(arg, ",")...)
		} else if len(strings.Split(arg, "|")) > 1 {
			arguments = append(arguments, strings.Split(arg, "|")...)
		} else {
			arguments = append(arguments, arg)
		}
	}

	for _, argument := range arguments {
		files, err := afero.Glob(fs, filepath.Clean(argument))
		if err != nil {
			return nil, err
		}

		for _, file := range files {
			if file != "." {
				asset := Asset{
					Name: filepath.Base(file),
					Path: file,
				}

				assets = append(assets, asset)
			}
		}
	}
	return &assets, nil
}

// Upload an asset to a GitHub release
func (a *Asset) Upload(release *Release, cli RepositoriesClient, id int64, msgs chan string, errs chan error, wg *sync.WaitGroup) {
	defer wg.Done()
	msgs <- fmt.Sprintf("uploading asset %v", a.Name)

	file, err := os.Open(a.Path)
	if err != nil {
		errs <- err
		return
	}
	defer file.Close()

	_, _, err = cli.UploadReleaseAsset(
		context.Background(),
		release.Slug.Owner,
		release.Slug.Name,
		id,
		&github.UploadOptions{
			Name: strings.ReplaceAll(a.Name, "/", "-"),
		},
		file,
	)
	if err != nil {
		errs <- err
		return
	}

	errs <- nil
}
