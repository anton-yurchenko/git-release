package changelog

import (
	"fmt"
	"strings"

	"github.com/pkg/errors"
	"github.com/spf13/afero"
)

// Changes represents changelog content for certain version
type Changes struct {
	File    string
	Version string
	Body    string
}

// Interface of 'Changes'
type Interface interface {
	ReadChanges(afero.Fs) error
	SetFile(string)
	GetFile() string
	GetBody() string
}

// ReadChanges loads section from changelog for a requested version
func (c *Changes) ReadChanges(fs afero.Fs) error {
	file, err := c.Read(fs)
	if err != nil {
		return err
	}

	margins := c.GetMargins(file)

	c.Body = strings.Join(GetContent(margins, file), "\n")

	if c.Body == "" {
		return errors.New(fmt.Sprintf("empty changelog for requested version: '%v'", c.Version))
	}

	return nil
}

// SetFile sets changelog filepath
func (c *Changes) SetFile(file string) {
	c.File = file
}

// GetFile returns changelog filepath
func (c *Changes) GetFile() string {
	return c.File
}

// GetBody returns changes body
func (c *Changes) GetBody() string {
	return c.Body
}
