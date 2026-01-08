package pkgsuggester

import (
	"context"
	"database/sql"

	"gorm.io/gorm"

	pkggorm "github.com/alphacodinggroup/ponti-backend/pkg/databases/sql/gorm"
)

type repoAdapter struct {
	repo *pkggorm.Repository
}

func NewPkggormAdapter(r *pkggorm.Repository) DB {
	return &repoAdapter{repo: r}
}

func (a *repoAdapter) WithContext(ctx context.Context) *miniGorm {
	return &miniGorm{inner: a.repo.Client().WithContext(ctx)}
}
func (a *repoAdapter) Exec(query string, args ...any) *miniGorm {
	return &miniGorm{inner: a.repo.Client().Exec(query, args...)}
}
func (a *repoAdapter) Raw(query string, args ...any) *miniGorm {
	return &miniGorm{inner: a.repo.Client().Raw(query, args...)}
}
func (a *repoAdapter) Scan(dest any) *miniGorm {
	return &miniGorm{inner: a.repo.Client().Scan(dest)}
}
func (a *repoAdapter) DB() (*sql.DB, error) {
	return a.repo.Client().DB()
}
func (a *repoAdapter) Error() error {
	return a.repo.Client().Error
}

type miniGorm struct {
	inner *gorm.DB
}

func (m *miniGorm) WithContext(ctx context.Context) *miniGorm {
	return &miniGorm{inner: m.inner.WithContext(ctx)}
}

func (m *miniGorm) Exec(query string, args ...any) *miniGorm {
	return &miniGorm{inner: m.inner.Exec(query, args...)}
}

func (m *miniGorm) Raw(query string, args ...any) *miniGorm {
	return &miniGorm{inner: m.inner.Raw(query, args...)}
}

func (m *miniGorm) Scan(dest any) *miniGorm {
	return &miniGorm{inner: m.inner.Scan(dest)}
}

func (m *miniGorm) DB() (*sql.DB, error) {
	return m.inner.DB()
}

func (m *miniGorm) Error() error {
	return m.inner.Error
}
