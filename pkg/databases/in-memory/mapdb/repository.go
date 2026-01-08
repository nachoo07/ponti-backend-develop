package pkgmapdb

import (
	"sync"

)

type Repository interface {
	GetDb() map[string]any
}


var (
	instance Repository
	once     sync.Once
)

type service struct {
	db map[string]any
}

func newRepository() Repository {
	once.Do(func() {
		instance = &service{
			db: make(map[string]any),
		}
	})
	return instance
}

func (c *service) GetDb() map[string]any {
	return c.db
}
