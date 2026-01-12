package pkgpdf

import (
	"fmt"
	"math"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"time"

	"github.com/jung-kurt/gofpdf"
)

// Repository (estilo pkggorm) mantiene el cliente PDF y la config.
// **Solo exporta PDF**.
type Repository struct {
	client  *gofpdf.Fpdf
	address string
	config  ConfigPort
}

// newRepository crea y conecta el cliente PDF.
func newRepository(c ConfigPort) (*Repository, error) {
	repo := &Repository{config: c}
	if err := repo.Connect(c); err != nil {
		return nil, fmt.Errorf("failed to initialize Repository: %w", err)
	}
	return repo, nil
}

// Connect crea un documento PDF nuevo con la configuración base.
func (r *Repository) Connect(config ConfigPort) error {
	p := gofpdf.New(config.GetOrientation(), "mm", config.GetPageSize(), "")
	p.SetMargins(config.GetMarginLeft(), config.GetMarginTop(), config.GetMarginRight())
	p.SetAutoPageBreak(true, config.GetMarginBottom())

	// UTF-8 fonts opcionales
	if config.GetUseUTF8() && config.GetFontRegularPath() != "" {
		p.AddUTF8Font(config.GetFontFamily(), "", config.GetFontRegularPath())
		if b := config.GetFontBoldPath(); b != "" {
			p.AddUTF8Font(config.GetFontFamily(), "B", b)
		}
	}
	p.SetFont(config.GetFontFamily(), "", config.GetFontSize())

	// Metadatos del documento
	if t := config.GetTitle(); t != "" {
		p.SetTitle(t, true)
	}
	if a := config.GetAuthor(); a != "" {
		p.SetAuthor(a, true)
	}
	if s := config.GetSubject(); s != "" {
		p.SetSubject(s, true)
	}
	if k := config.GetKeywords(); k != "" {
		p.SetKeywords(k, true)
	}
	if config.GetPageNumbers() {
		p.AliasNbPages("")
		p.SetFooterFunc(func() {
			p.SetY(-12)
			p.SetFont(config.GetFontFamily(), "", config.GetFontSize()-2)
			p.CellFormat(0, 10, fmt.Sprintf("Page %d/{nb}", p.PageNo()), "", 0, "C", false, 0, "")
		})
	}
	r.client = p
	r.address = config.GetFilePath()
	r.config = config
	return nil
}

// Client retorna el *gofpdf.Fpdf actual.
func (r *Repository) Client() *gofpdf.Fpdf { return r.client }

// Address retorna el path de salida.
func (r *Repository) Address() string { return r.address }

// ExportTable serializa un slice/array de structs o maps como tabla y guarda el PDF en Address().
func (r *Repository) ExportTable(data any) error {
	if r.client == nil {
		return fmt.Errorf("client not connected")
	}
	if dir := filepath.Dir(r.address); dir != "" && dir != "." {
		_ = os.MkdirAll(dir, 0o755)
	}
	if err := r.render(r.client, data); err != nil {
		return err
	}
	return r.client.OutputFileAndClose(r.address)
}

// --------- Internos ---------

// render aplica título, header y filas (wrapping) en el PDF ya configurado.
func (r *Repository) render(p *gofpdf.Fpdf, data any) error {
	headers, rows, err := r.buildRows(data)
	if err != nil {
		return err
	}
	if len(headers) == 0 {
		return fmt.Errorf("no headers derived from data")
	}

	p.AddPage()

	// Título opcional.
	if t := r.config.GetTitle(); t != "" {
		p.SetFont(r.config.GetFontFamily(), "B", r.config.GetFontSize()+2)
		p.CellFormat(0, 8, t, "0", 1, "L", false, 0, "")
		p.Ln(2)
		p.SetFont(r.config.GetFontFamily(), "", r.config.GetFontSize())
	}

	// Anchos de columna.
	pageW, pageH := p.GetPageSize()
	lm, _, rm, _ := p.GetMargins()
	bm := r.config.GetMarginBottom()
	usableW := pageW - lm - rm
	colW := r.columnWidths(p, headers, rows, usableW)

	// Header.
	rowH := 8.0
	drawHeader := func() {
		p.SetFont(r.config.GetFontFamily(), "B", r.config.GetFontSize())
		for i, h := range headers {
			p.CellFormat(colW[i], rowH, h, "1", 0, "C", false, 0, "")
		}
		p.Ln(-1)
		p.SetFont(r.config.GetFontFamily(), "", r.config.GetFontSize())
	}
	drawHeader()

	// Rows con wrapping simple.
	for _, row := range rows {
		maxH := rowH
		for i, txt := range row {
			lines := p.SplitLines([]byte(txt), colW[i]-2)
			h := float64(len(lines)) * (rowH - 2)
			if h > maxH {
				maxH = math.Max(rowH, h)
			}
		}
		// salto de página si no entra la fila
		_, y := p.GetXY()
		if y+maxH > (pageH - bm) {
			p.AddPage()
			if r.config.GetHeaderEveryPage() {
				drawHeader()
			}
		}
		x, y := p.GetXY()
		for i, txt := range row {
			p.Rect(x, y, colW[i], maxH, "D")
			p.SetXY(x+1, y+1)
			p.MultiCell(colW[i]-2, rowH-2, txt, "", "L", false)
			x += colW[i]
		}
		p.SetXY(lm, y+maxH)
	}
	return nil
}

func (r *Repository) columnWidths(p *gofpdf.Fpdf, headers []string, rows [][]string, usableW float64) []float64 {
	n := len(headers)
	if ws := r.config.GetColumnWidths(); len(ws) == n {
		return ws
	}
	w := make([]float64, n)
	pad := 4.0
	for i, h := range headers {
		w[i] = p.GetStringWidth(h) + pad
	}
	for _, row := range rows {
		for i, cell := range row {
			cw := p.GetStringWidth(cell) + pad
			if cw > w[i] {
				w[i] = cw
			}
		}
	}
	sum := 0.0
	for _, v := range w {
		sum += v
	}
	if sum > usableW {
		scale := usableW / sum
		for i := range w {
			w[i] *= scale
		}
	}
	return w
}

// buildRows: soporta slice/array de structs (tags `pdf:"Header"` o `json`) o de maps.
func (r *Repository) buildRows(data any) ([]string, [][]string, error) {
	v := reflect.ValueOf(data)
	if v.Kind() != reflect.Slice && v.Kind() != reflect.Array {
		return nil, nil, fmt.Errorf("data must be slice or array")
	}
	if v.Len() == 0 {
		return nil, nil, fmt.Errorf("data slice is empty")
	}
	el := v.Index(0)
	if el.Kind() == reflect.Ptr {
		el = el.Elem()
	}
	switch el.Kind() {
	case reflect.Struct:
		return r.rowsFromStructs(v)
	case reflect.Map:
		return r.rowsFromMaps(v)
	default:
		return nil, nil, fmt.Errorf("unsupported element kind: %s", el.Kind())
	}
}

func (r *Repository) rowsFromStructs(v reflect.Value) ([]string, [][]string, error) {
	t := v.Index(0).Type()
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	nf := t.NumField()
	idx := make([]int, 0, nf)
	headers := make([]string, 0, nf)
	for i := 0; i < nf; i++ {
		f := t.Field(i)
		if f.PkgPath != "" { // no exportado
			continue
		}
		if tag := f.Tag.Get("pdf"); tag == "-" {
			continue
		}
		name := f.Tag.Get("pdf")
		if name == "" {
			name = f.Tag.Get("json")
		}
		if name == "" || name == "-" {
			name = f.Name
		}
		headers = append(headers, name)
		idx = append(idx, i)
	}
	rows := make([][]string, v.Len())
	for rIdx := 0; rIdx < v.Len(); rIdx++ {
		row := make([]string, len(idx))
		item := v.Index(rIdx)
		if item.Kind() == reflect.Ptr {
			item = item.Elem()
		}
		for c, fi := range idx {
			row[c] = formatVal(item.Field(fi))
		}
		rows[rIdx] = row
	}
	return headers, rows, nil
}

func (r *Repository) rowsFromMaps(v reflect.Value) ([]string, [][]string, error) {
	keySet := map[string]struct{}{}
	for i := 0; i < v.Len(); i++ {
		it := v.Index(i)
		if it.Kind() == reflect.Ptr {
			it = it.Elem()
		}
		for _, k := range it.MapKeys() {
			keySet[k.String()] = struct{}{}
		}
	}
	headers := make([]string, 0, len(keySet))
	for k := range keySet {
		headers = append(headers, k)
	}
	sort.Strings(headers)

	rows := make([][]string, v.Len())
	for rIdx := 0; rIdx < v.Len(); rIdx++ {
		row := make([]string, len(headers))
		it := v.Index(rIdx)
		if it.Kind() == reflect.Ptr {
			it = it.Elem()
		}
		for c, hk := range headers {
			val := it.MapIndex(reflect.ValueOf(hk))
			row[c] = formatVal(val)
		}
		rows[rIdx] = row
	}
	return headers, rows, nil
}

func formatVal(v reflect.Value) string {
	if !v.IsValid() {
		return ""
	}
	if v.Kind() == reflect.Interface || v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return ""
		}
		v = v.Elem()
	}
	switch v.Kind() {
	case reflect.String:
		return v.String()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return fmt.Sprintf("%d", v.Int())
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return fmt.Sprintf("%d", v.Uint())
	case reflect.Float32, reflect.Float64:
		return fmt.Sprintf("%g", v.Float())
	case reflect.Bool:
		if v.Bool() {
			return "true"
		}
		return "false"
	case reflect.Struct:
		if t, ok := v.Interface().(time.Time); ok {
			return t.Format(time.RFC3339)
		}
	}
	return fmt.Sprintf("%v", v.Interface())
}
