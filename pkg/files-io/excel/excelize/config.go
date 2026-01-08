// Package: pkgexcel (estilo pkggorm) — 3 archivos: config.go, bootstrap.go, service.go
// File: config.go
package pkgexcel

import (
	"fmt"
)

// ConfigPort define los getters y la validación, igual al patrón de pkggorm.
type ConfigPort interface {
	GetFilePath() string
	GetSheet() string
	GetDateFormat() string
	GetWriteHeader() bool
	GetColumnWidths() []float64
	Validate() error
}

// Config concreta (como en pkggorm)
type Config struct {
	filePath     string
	sheet        string
	dateFormat   string
	writeHeader  bool
	columnWidths []float64
}

// newConfig crea una Config como hace pkggorm
func newConfig(filePath, sheet, dateFormat string, writeHeader bool, widths []float64) *Config {
	return &Config{
		filePath:     filePath,
		sheet:        sheet,
		dateFormat:   dateFormat,
		writeHeader:  writeHeader,
		columnWidths: widths,
	}
}

// Getters (mismo estilo pkggorm)
func (c *Config) GetFilePath() string        { return c.filePath }
func (c *Config) GetSheet() string           { return c.sheet }
func (c *Config) GetDateFormat() string      { return c.dateFormat }
func (c *Config) GetWriteHeader() bool       { return c.writeHeader }
func (c *Config) GetColumnWidths() []float64 { return c.columnWidths }

// Validate (con defaults, como en pkggorm)
func (c *Config) Validate() error {
	if c.filePath == "" {
		return fmt.Errorf("file path is required")
	}
	if c.sheet == "" {
		c.sheet = "Sheet1"
	}
	if c.dateFormat == "" {
		c.dateFormat = "2006-01-02"
	}
	// El default ya se resolvió en Bootstrap con tri‑state.
	return nil
}

// Interface assertion guards to ensure Config satisfies ConfigPort at compile-time.
var _ ConfigPort = (*Config)(nil)
