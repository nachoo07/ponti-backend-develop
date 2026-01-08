package pkgswagger

import "fmt"

type config struct {
	title       string
	description string
	version     string
	host        string
	basePath    string
	schemes     []string
	enabled     bool
}

func newConfig(title, description, version, host, basePath string, schemes []string, enabled bool) *config {
	return &config{
		title:       title,
		description: description,
		version:     version,
		host:        host,
		basePath:    basePath,
		schemes:     schemes,
		enabled:     enabled,
	}
}

func (c *config) GetTitle() string       { return c.title }
func (c *config) GetDescription() string { return c.description }
func (c *config) GetVersion() string     { return c.version }
func (c *config) GetHost() string        { return c.host }
func (c *config) GetBasePath() string    { return c.basePath }
func (c *config) GetSchemes() []string   { return c.schemes }
func (c *config) IsEnabled() bool        { return c.enabled }

func (c *config) Validate() error {
	if c.title == "" {
		return fmt.Errorf("swagger title is not configured")
	}
	if c.version == "" {
		return fmt.Errorf("swagger version is not configured")
	}
	if len(c.schemes) == 0 {
		c.schemes = []string{"http"}
	}
	if c.host == "" {
		c.host = "localhost:8080"
	}
	if c.basePath == "" {
		c.basePath = "/"
	}
	return nil
}
