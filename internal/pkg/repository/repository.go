package repository

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/pkg/errors"
)

// Repository represents a local git project repository
type Repository struct {
	Owner      string
	Project    string
	CommitHash string
	Tag        string
}

// Interface of 'Repository'
type Interface interface {
	ReadTag(*string, bool) error
	ReadCommitHash()
	ReadProjectName() error
	GetOwner() string
	GetProject() string
	GetTag() *string
	GetCommitHash() *string
}

// ReadTag sets tag to the receiver and sem.ver parsed version to provided parameter
func (r *Repository) ReadTag(version *string, allowPrefix bool) error {
	ref := os.Getenv("GITHUB_REF")

	semver := "[v]?(?P<major>0|[1-9]\\d*)\\.(?P<minor>0|[1-9]\\d*)\\.(?P<patch>0|[1-9]\\d*)(?:(?P<sep1>-)(?P<prerelease>(?:0|[1-9]\\d*|\\d*[a-zA-Z-][0-9a-zA-Z-]*)(?:\\.(?:0|[1-9]\\d*|\\d*[a-zA-Z-][0-9a-zA-Z-]*))*))?(?:(?P<sep2>\\+)(?P<buildmetadata>[0-9a-zA-Z-]+(?:\\.[0-9a-zA-Z-]+)*))?"

	var expression string
	if allowPrefix {
		expression = fmt.Sprintf("^refs/tags/(?P<prefix>.*?)%v$", semver)
	} else {
		expression = fmt.Sprintf("^refs/tags/%v$", semver)
	}

	regex := regexp.MustCompile(expression)

	if regex.MatchString(ref) {
		refs := strings.Split(ref, "/")
		r.Tag = strings.Join(refs[2:], "/")

		if allowPrefix {
			*version = regex.ReplaceAllString(ref, "${2}.${3}.${4}${5}${6}${7}${8}")
		} else {
			*version = regex.ReplaceAllString(ref, "${1}.${2}.${3}${4}${5}${6}${7}${8}")
		}

		return nil
	}

	return errors.New(fmt.Sprintf("malformed env.var 'GITHUB_REF' (control tag prefix via env.var 'ALLOW_TAG_PREFIX'): expected to match regex '%s', got '%v'", expression, ref))
}

// ReadCommitHash sets current commit hash
func (r *Repository) ReadCommitHash() {
	r.CommitHash = os.Getenv("GITHUB_SHA")
}

// ReadProjectName sets parsed owner and project names
func (r *Repository) ReadProjectName() error {
	o := os.Getenv("GITHUB_REPOSITORY")
	if o == "" {
		return errors.New("env.var 'GITHUB_REPOSITORY' is empty or not defined")
	}

	expression := "^(?P<owner>[\\w,\\-,\\_\\.]+)\\/(?P<repo>[\\w\\,\\-\\_\\.]+)$"
	regex := regexp.MustCompile(expression)

	if regex.MatchString(o) {
		r.Owner = strings.Split(o, "/")[0]
		r.Project = strings.Split(o, "/")[1]

		return nil
	}

	return errors.New(fmt.Sprintf("malformed env.var 'GITHUB_REPOSITORY': expected to match regex '%v', got '%v'", expression, o))
}

// GetOwner returns project owner
func (r *Repository) GetOwner() string {
	return r.Owner
}

// GetProject returns project name
func (r *Repository) GetProject() string {
	return r.Project
}

// GetTag returns current tag
func (r *Repository) GetTag() *string {
	return &r.Tag
}

// GetCommitHash returns current commit hash
func (r *Repository) GetCommitHash() *string {
	return &r.CommitHash
}
