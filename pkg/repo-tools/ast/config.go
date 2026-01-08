package pkgast

import (
	"fmt"

	defs "github.com/alphacodinggroup/ponti-backend/pkg/repo-tools/ast/defs"
)

type config struct {
	AnalyzePath string
}

func newConfig(analyzePath string) defs.Config {
	return &config{
		AnalyzePath: analyzePath,
	}
}

func (c *config) GetAnalyzePath() string {
	return c.AnalyzePath
}

func (c *config) SetAnalyzePath(analyzePath string) {
	c.AnalyzePath = analyzePath
}

func (c *config) Validate() error {
	if c.AnalyzePath == "" {
		return fmt.Errorf("analyze path is not configured")
	}
	return nil
}
