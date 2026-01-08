package pkggogit

import (
	"fmt"

	defs "github.com/alphacodinggroup/ponti-backend/pkg/repo-tools/go-git/v5/defs"
)

type config struct {
	RepoURL    string
	RepoPath   string
	RepoBranch string
}

func newConfig(repoURL, repoPath, repoBranch string) defs.Config {
	return &config{
		RepoURL:    repoURL,
		RepoPath:   repoPath,
		RepoBranch: repoBranch,
	}
}

func (c *config) GetRepoURL() string {
	return c.RepoURL
}

func (c *config) SetRepoURL(repoURL string) {
	c.RepoURL = repoURL
}

func (c *config) GetRepoPath() string {
	return c.RepoPath
}

func (c *config) SetRepoPath(repoPath string) {
	c.RepoPath = repoPath
}

func (c *config) GetRepoBranch() string {
	return c.RepoBranch
}

func (c *config) SetRepoBranch(repoBranch string) {
	c.RepoBranch = repoBranch
}

func (c *config) Validate() error {
	if c.RepoURL == "" {
		return fmt.Errorf("GIT_REPO_URL is not configured")
	}
	if c.RepoPath == "" {
		return fmt.Errorf("GIT_REPO_PATH is not configured")
	}
	return nil
}
