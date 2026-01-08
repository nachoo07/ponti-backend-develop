package main

import (
	"log"
	"os"
	"path/filepath"
	"time"

	pkgexcel "github.com/alphacodinggroup/ponti-backend/pkg/files-io/excel/excelize"
)

type row struct {
	ID        int       `excel:"ID"`
	Name      string    `excel:"Nombre"`
	Price     float64   `excel:"Precio"`
	CreatedAt time.Time `excel:"Creado"`
}

func main() {
	out := filepath.Join(os.TempDir(), "excel_demo.xlsx")

	// Opcionales por ENV: EXCEL_SHEET, EXCEL_DATE_FORMAT, EXCEL_WRITE_HEADER, EXCEL_COLUMN_WIDTHS
	_ = os.Setenv("EXCEL_SHEET", "Demo")
	_ = os.Setenv("EXCEL_DATE_FORMAT", "2006-01-02 15:04")

	svc, err := pkgexcel.Bootstrap(out, "", "", nil, nil)
	if err != nil {
		log.Fatalf("bootstrap error: %v", err)
	}
	defer func() { _ = svc.Close() }()

	data := []row{
		{ID: 1, Name: "Producto A", Price: 12.5, CreatedAt: time.Now()},
		{ID: 2, Name: "Producto B", Price: 99.99, CreatedAt: time.Now().Add(-24 * time.Hour)},
		{ID: 3, Name: "Producto C", Price: 0, CreatedAt: time.Now().Add(-48 * time.Hour)},
	}

	if err := svc.Export(data); err != nil {
		log.Fatalf("export error: %v", err)
	}
	log.Printf("Excel generado en: %s", out)
}
