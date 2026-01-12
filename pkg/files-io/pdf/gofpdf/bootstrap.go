package pkgpdf

import (
	"os"
	"strconv"
	"strings"
)

// Bootstrap inicializa el repositorio PDF.
// ENV soportados: PDF_FILE_PATH, PDF_ORIENTATION, PDF_PAGE_SIZE, PDF_MARGIN_LEFT, PDF_MARGIN_TOP,
// PDF_MARGIN_RIGHT, PDF_MARGIN_BOTTOM, PDF_FONT_FAMILY, PDF_FONT_SIZE, PDF_TITLE,
// PDF_AUTHOR, PDF_SUBJECT, PDF_KEYWORDS, PDF_USE_UTF8, PDF_FONT_REGULAR_PATH, PDF_FONT_BOLD_PATH,
// PDF_PAGE_NUMBERS, PDF_HEADER_EVERY_PAGE, PDF_FOOTER_EVERY_PAGE, PDF_HEADER_HEIGHT, PDF_FOOTER_HEIGHT.
func Bootstrap(
	filePath, orientation, pageSize, fontFamily, title string,
	fontSize, ml, mt, mr float64,
	columnWidths []float64,
) (*Repository, error) {

	if filePath == "" {
		filePath = os.Getenv("PDF_FILE_PATH")
	}
	if orientation == "" {
		orientation = os.Getenv("PDF_ORIENTATION")
	}
	if pageSize == "" {
		pageSize = os.Getenv("PDF_PAGE_SIZE")
	}
	if fontFamily == "" {
		fontFamily = os.Getenv("PDF_FONT_FAMILY")
	}
	if title == "" {
		title = os.Getenv("PDF_TITLE")
	}
	if fontSize == 0 {
		if v := os.Getenv("PDF_FONT_SIZE"); v != "" {
			if f, err := strconv.ParseFloat(v, 64); err == nil {
				fontSize = f
			}
		}
	}
	if ml == 0 {
		if v := os.Getenv("PDF_MARGIN_LEFT"); v != "" {
			if f, err := strconv.ParseFloat(v, 64); err == nil {
				ml = f
			}
		}
	}
	if mt == 0 {
		if v := os.Getenv("PDF_MARGIN_TOP"); v != "" {
			if f, err := strconv.ParseFloat(v, 64); err == nil {
				mt = f
			}
		}
	}
	if mr == 0 {
		if v := os.Getenv("PDF_MARGIN_RIGHT"); v != "" {
			if f, err := strconv.ParseFloat(v, 64); err == nil {
				mr = f
			}
		}
	}
	mb := 0.0
	if v := os.Getenv("PDF_MARGIN_BOTTOM"); v != "" {
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			mb = f
		}
	}

	author := os.Getenv("PDF_AUTHOR")
	subject := os.Getenv("PDF_SUBJECT")
	keywords := os.Getenv("PDF_KEYWORDS")

	useUTF8 := false
	if v := strings.TrimSpace(os.Getenv("PDF_USE_UTF8")); v != "" {
		if b, err := strconv.ParseBool(v); err == nil {
			useUTF8 = b
		}
	}
	fontRegularPath := os.Getenv("PDF_FONT_REGULAR_PATH")
	fontBoldPath := os.Getenv("PDF_FONT_BOLD_PATH")

	headerEveryPage := false
	if v := strings.TrimSpace(os.Getenv("PDF_HEADER_EVERY_PAGE")); v != "" {
		if b, err := strconv.ParseBool(v); err == nil {
			headerEveryPage = b
		}
	}
	pageNumbers := false
	if v := strings.TrimSpace(os.Getenv("PDF_PAGE_NUMBERS")); v != "" {
		if b, err := strconv.ParseBool(v); err == nil {
			pageNumbers = b
		}
	}

	cfg := newConfig(
		filePath, orientation, pageSize, fontFamily, title, author, subject, keywords,
		fontSize, ml, mt, mr, mb,
		columnWidths,
		useUTF8,
		fontRegularPath, fontBoldPath,
		headerEveryPage, pageNumbers,
	)
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	return newRepository(cfg)
}
