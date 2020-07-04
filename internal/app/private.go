package app

import (
	"path/filepath"
	"strings"

	"github.com/anton-yurchenko/git-release/internal/pkg/asset"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/afero"
)

// GetAssets returns validated assets supplied via 'args'
func GetAssets(fs afero.Fs, args []string) []asset.Asset {
	assets := make([]asset.Asset, 0)
	arguments := make([]string, 0)

	for _, arg := range args {
		// split arguments by space, new line, comma, pipe
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
			log.Fatal(err)
		}

		for _, file := range files {
			if file != "." {
				asset := asset.Asset{
					Name: filepath.Base(file),
					Path: file,
				}

				assets = append(assets, asset)
			}
		}
	}
	return assets
}

// IsExists validates whether a file exists and returns the result as a bool
func IsExists(file string, fs afero.Fs) (bool, error) {
	return afero.Exists(fs, file)
}
