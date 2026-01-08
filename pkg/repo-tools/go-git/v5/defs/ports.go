package defs

import (
	"github.com/go-git/go-git/v5"
)

// GitClient es la interfaz que define las operaciones que debe soportar un cliente Git.
type Client interface {
	GetRepository() *git.Repository
	PullLatest() error
	GetFiles(files []string, extension string) ([]string, error)
	GetFileAuthor(file string) (string, error)
	GetCommitID(file string) (string, error)
	GetRepo(repoPath string) (*git.Repository, error)
}

type Config interface {
	GetRepoURL() string
	SetRepoURL(string)
	GetRepoPath() string
	SetRepoPath(string)
	GetRepoBranch() string
	SetRepoBranch(string)
	Validate() error
}
