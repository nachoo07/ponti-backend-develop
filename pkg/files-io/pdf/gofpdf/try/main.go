package main

import (
	"log"
	"os"
	"path/filepath"
	"time"

	pkgpdf "github.com/alphacodinggroup/ponti-backend/pkg/files-io/pdf/gofpdf"
)

type row struct {
	Name      string    `pdf:"Nombre"`
	Age       int       `pdf:"Edad"`
	CreatedAt time.Time `pdf:"Creado"`
}

func main() {
	// Archivo de salida
	out := filepath.Join(os.TempDir(), "pdf_demo.pdf")

	// Configuración opcional por ENV
	_ = os.Setenv("PDF_PAGE_NUMBERS", "true")
	_ = os.Setenv("PDF_HEADER_EVERY_PAGE", "true")
	_ = os.Setenv("PDF_MARGIN_BOTTOM", "15")

	repo, err := pkgpdf.Bootstrap(
		out,        // filePath
		"P",        // orientation
		"A4",       // pageSize
		"Arial",    // fontFamily
		"Demo PDF", // title
		12,         // fontSize
		12,         // margin left
		12,         // margin top
		12,         // margin right
		nil,        // column widths (auto)
	)
	if err != nil {
		log.Fatalf("bootstrap error: %v", err)
	}

	// Datos de prueba
	data := make([]row, 0, 120)
	for i := 0; i < 120; i++ {
		data = append(data, row{
			Name:      "Row #" + itoa(i) + " – Long text to test wrapping",
			Age:       20 + (i % 50),
			CreatedAt: time.Now().Add(time.Duration(-i) * time.Hour),
		})
	}

	if err := repo.ExportTable(data); err != nil {
		log.Fatalf("export error: %v", err)
	}
	log.Printf("PDF generado en: %s", out)
}

func itoa(i int) string {
	// Función auxiliar para convertir int a string
	return fmtInt(i)
}

// simple int->string conversion without strconv (to keep minimum imports)
func fmtInt(i int) string {
	if i == 0 {
		return "0"
	}
	neg := false
	if i < 0 {
		neg = true
		i = -i
	}
	var digits [20]byte
	pos := len(digits)
	for i > 0 {
		pos--
		digits[pos] = byte('0' + (i % 10))
		i /= 10
	}
	if neg {
		pos--
		digits[pos] = '-'
	}
	return string(digits[pos:])
}
