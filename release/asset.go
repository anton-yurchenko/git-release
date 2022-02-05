package release

import (
	"context"
	"fmt"
	"math"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/google/go-github/github"
	"github.com/pkg/errors"
	"github.com/spf13/afero"

	log "github.com/sirupsen/logrus"
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
func (a *Asset) Upload(release *Release, cli RepositoriesClient, id int64, errs chan error, wg *sync.WaitGroup) {
	defer wg.Done()
	log.WithField("asset", a.Name).Info("uploading asset")

	maxRetries := 4
	for i := 1; i <= maxRetries; i++ {
		err := a.uploadHandler(
			release,
			cli,
			id,
			i == maxRetries,
		)
		if err == nil {
			errs <- nil
			break
		} else if strings.Contains(err.Error(), "error opening a file") {
			errs <- err
			return
		}

		if i == maxRetries {
			errs <- errors.New(fmt.Sprintf("maximum attempts reached uploading asset: %v", a.Name))
			break
		}

		log.WithField("asset", a.Name).Warn(err.Error())

		delay := math.Pow(3, float64(i+1))
		log.WithField("asset", a.Name).Infof("retrying (%v/%v) uploading asset in %v seconds", i+1, maxRetries, delay)
		time.Sleep(time.Duration(delay) * time.Second)
	}
}

func (a *Asset) uploadHandler(release *Release, cli RepositoriesClient, id int64, lastTry bool) error {
	file, err := os.Open(a.Path)
	if err != nil {
		return errors.Wrap(err, "error opening a file")
	}

	_, res, err := cli.UploadReleaseAsset(
		context.Background(),
		release.Slug.Owner,
		release.Slug.Name,
		id,
		&github.UploadOptions{
			Name: strings.ReplaceAll(a.Name, "/", "-"),
		},
		file,
	)

	_ = file.Close()

	if err != nil {
		log.WithField("asset", a.Name).Warnf("error uploading asset: %v", err.Error())

		if !lastTry && (res.StatusCode == http.StatusBadGateway || res.StatusCode == http.StatusUnprocessableEntity) {
			rel, _, err := cli.GetReleaseByTag(
				context.Background(),
				release.Slug.Owner,
				release.Slug.Name,
				release.Reference.Tag,
			)
			if err != nil {
				return errors.Wrap(err, "error retrieving release")
			}

			for _, s := range rel.Assets {
				if *s.Name == strings.ReplaceAll(a.Name, "/", "-") {
					_, err = cli.DeleteReleaseAsset(
						context.Background(),
						release.Slug.Owner,
						release.Slug.Name,
						*s.ID,
					)
					if err != nil {
						return errors.Wrap(err, "error deleting ghost release asset")
					}

					return errors.New("ghost release asset deleted")
				}
			}

			return errors.New("ghost release asset not found")
		}

		return err
	}

	return nil
}
