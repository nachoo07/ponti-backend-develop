package pkgsqlite

import (
	"fmt"
)

type Config interface {
	Validate() error
	DNS() string
	GetDBPath() string
	GetInMemory() bool
}

type config struct {
	DBPath   string
	InMemory bool
}

func newConfig(dbPath string, inMemory bool) Config {
	return &config{
		DBPath:   dbPath,
		InMemory: inMemory,
	}
}

func (c *config) DNS() string {
	if c.InMemory {
		return "file::memory:?cache=shared"
	}
	return fmt.Sprintf("file:%s?cache=shared", c.DBPath)
}

func (c *config) GetDBPath() string {
	return c.DBPath
}

func (c *config) GetInMemory() bool {
	return c.InMemory
}

func (c *config) Validate() error {
	if !c.InMemory && c.DBPath == "" {
		return fmt.Errorf("SQLITE_DB_PATH environment variable is empty and not using in-memory database")
	}
	return nil
}
