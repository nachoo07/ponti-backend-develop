package pkgexcel

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// Bootstrap inicializa el repositorio con parámetros o ENV.
// ENV soportados (como en pkggorm)
// Returns error if after resolving ENV critical values like filePath or sheet are missing.
func Bootstrap(filePath, sheet, dateFormat string, writeHeader *bool, columnWidths []float64) (*Service, error) {
	if filePath == "" {
		filePath = os.Getenv("EXCEL_FILE_PATH")
	}
	if sheet == "" {
		sheet = os.Getenv("EXCEL_SHEET")
	}
	if dateFormat == "" {
		if v := os.Getenv("EXCEL_DATE_FORMAT"); v != "" {
			dateFormat = v
		}
	}
	// Tri‑state para writeHeader: ENV → parámetro → default
	if writeHeader == nil {
		if v := os.Getenv("EXCEL_WRITE_HEADER"); v != "" {
			if b, err := strconv.ParseBool(v); err == nil {
				writeHeader = &b
			}
		}
	}
	if writeHeader == nil {
		def := true
		writeHeader = &def
	}

	// Parsear anchos de columna desde ENV
	if len(columnWidths) == 0 {
		if v := os.Getenv("EXCEL_COLUMN_WIDTHS"); v != "" {
			parts := strings.Split(v, ",")
			ws := make([]float64, 0, len(parts))
			for _, p := range parts {
				if f, err := strconv.ParseFloat(strings.TrimSpace(p), 64); err == nil {
					ws = append(ws, f)
				}
			}
			if len(ws) > 0 {
				columnWidths = ws
			}
		}
	}

	// Ensure destination directory
	if dir := filepath.Dir(filePath); dir != "." && dir != "" {
		_ = os.MkdirAll(dir, 0o755)
	}

	// Validar parámetros obligatorios
	if filePath == "" {
		return nil, fmt.Errorf("excel bootstrap: file path is required (EXCEL_FILE_PATH)")
	}
	if sheet == "" {
		return nil, fmt.Errorf("excel bootstrap: sheet name is required (EXCEL_SHEET)")
	}

	cfg := newConfig(filePath, sheet, dateFormat, *writeHeader, columnWidths)
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	return newService(cfg)
}
