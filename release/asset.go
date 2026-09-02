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

	"github.com/google/go-github/v78/github"
	"github.com/pkg/errors"
	"github.com/spf13/afero"

	log "github.com/sirupsen/logrus"
)

// GetAssets expands the configured paths and globs into release assets.
//
// NOTE: patterns arrive already split, one per line. v6 received the list on
// argv and guessed the separator - splitting on whichever of space, newline,
// comma or pipe matched first - which made a path containing a space
// inexpressible, silently discarded the remainder of a comma-and-space
// separated list, and produced a different asset set through the container than
// through the JavaScript wrapper.
func GetAssets(fs afero.Fs, patterns []string) (*[]Asset, error) {
	assets := make([]Asset, 0)

	for _, pattern := range patterns {
		files, err := afero.Glob(fs, filepath.Clean(pattern))
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

// retryDelay is the backoff before retry attempt i (1-based).
//
// NOTE: a package-level variable so tests can collapse it. The suite otherwise
// sleeps 9 + 27 + 81 real seconds for every retry-exhausting case, which made
// the upload tests too slow to assert on properly.
var retryDelay = func(i int) time.Duration {
	return time.Duration(math.Pow(3, float64(i+1))) * time.Second
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

		delay := retryDelay(i)
		log.WithField("asset", a.Name).Infof("retrying (%v/%v) uploading asset in %v seconds", i+1, maxRetries, delay.Seconds())
		time.Sleep(delay)
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

		if !lastTry &&
			res != nil && res.Response != nil &&
			(res.StatusCode == http.StatusBadGateway || res.StatusCode == http.StatusUnprocessableEntity) {
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
