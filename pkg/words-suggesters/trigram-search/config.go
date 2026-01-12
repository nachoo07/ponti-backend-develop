package pkgsuggester

import (
	"context"
	"database/sql"
	"fmt"
	"regexp"
	"time"
)

const (
	defaultLimit     = 10
	defaultThreshold = 0.3
	defaultSchema    = "public"
)

var validIdentifier = regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]*$`)

type DB interface {
	WithContext(ctx context.Context) *miniGorm
	Exec(query string, args ...any) *miniGorm
	Raw(query string, args ...any) *miniGorm
	Scan(dest any) *miniGorm
	DB() (*sql.DB, error)
	Error() error
}

type Option func(*Config)

// Match modes for suggestion strategy.
type MatchMode int

const (
	ModePrefix MatchMode = iota
	ModeTrigram
	ModeHybrid
)

// Count strategies for total rows.
type CountMode int

const (
	CountExact CountMode = iota
	CountPrefixThenFuzzyIfNeeded
	CountApproxNone
)

type Config struct {
	DB               DB
	Limit            int
	Threshold        float64
	logger           Logger
	Schema           string
	IDColumn         string
	Mode             MatchMode
	Count            CountMode
	QueryTimeout     time.Duration
	EnsureExtensions bool
}

// Opciones de configuración
func WithDB(db DB) Option                     { return func(c *Config) { c.DB = db } }
func WithLimit(n int) Option                  { return func(c *Config) { c.Limit = n } }
func WithThreshold(t float64) Option          { return func(c *Config) { c.Threshold = t } }
func WithLogger(l Logger) Option              { return func(c *Config) { c.logger = l } }
func WithSchema(s string) Option              { return func(c *Config) { c.Schema = s } }
func WithIDColumn(s string) Option            { return func(c *Config) { c.IDColumn = s } }
func WithMode(m MatchMode) Option             { return func(c *Config) { c.Mode = m } }
func WithCountMode(cm CountMode) Option       { return func(c *Config) { c.Count = cm } }
func WithQueryTimeout(d time.Duration) Option { return func(c *Config) { c.QueryTimeout = d } }
func WithEnsureExtensions(b bool) Option      { return func(c *Config) { c.EnsureExtensions = b } }

func newConfig(opts ...Option) (*Config, error) {
	cfg := &Config{
		DB:               nil,
		Limit:            defaultLimit,
		Threshold:        defaultThreshold,
		logger:           noopLogger{},
		Schema:           defaultSchema,
		IDColumn:         "id",
		Mode:             ModeHybrid,
		Count:            CountExact,
		QueryTimeout:     2 * time.Second,
		EnsureExtensions: false,
	}
	for _, o := range opts {
		o(cfg)
	}
	if err := cfg.validate(); err != nil {
		return nil, err
	}
	return cfg, nil
}

func (c *Config) validate() error {
	if c.DB == nil {
		return fmt.Errorf("must provide a DB via WithDB")
	}
	if c.Limit <= 0 {
		return fmt.Errorf("limit must be > 0, got %d", c.Limit)
	}
	if c.Threshold < 0 || c.Threshold > 1 {
		return fmt.Errorf("threshold must be between 0 and 1, got %f", c.Threshold)
	}
	if c.Schema == "" || !validIdentifier.MatchString(c.Schema) {
		return fmt.Errorf("invalid schema name: %q", c.Schema)
	}
	if c.IDColumn == "" || !validIdentifier.MatchString(c.IDColumn) {
		return fmt.Errorf("invalid id column: %q", c.IDColumn)
	}
	return nil
}
