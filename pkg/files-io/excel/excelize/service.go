package pkgexcel

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"time"

	"database/sql"

	"github.com/shopspring/decimal"
	"github.com/xuri/excelize/v2"
)

type Service struct {
	client  *excelize.File
	address string
	config  ConfigPort
}

// newService inicializa un repositorio y conecta.
func newService(c ConfigPort) (*Service, error) {
	repo := &Service{config: c}
	if err := repo.Connect(c); err != nil {
		return nil, fmt.Errorf("failed to initialize Service: %w", err)
	}
	return repo, nil
}

// Connect abre o crea el archivo y garantiza la hoja.
func (r *Service) Connect(config ConfigPort) error {
	if err := r.createWorkbookIfNotExists(config); err != nil {
		return fmt.Errorf("failed to create workbook: %w", err)
	}
	f, err := excelize.OpenFile(config.GetFilePath())
	if err != nil {
		return fmt.Errorf("open workbook: %w", err)
	}
	r.client = f
	r.address = config.GetFilePath()
	// asegurar hoja
	idx, err := r.client.GetSheetIndex(config.GetSheet())
	if err != nil || idx < 0 {
		idx, err = r.client.NewSheet(config.GetSheet())
		if err != nil {
			return fmt.Errorf("create sheet %q: %w", config.GetSheet(), err)
		}
	}
	r.client.SetActiveSheet(idx)
	return nil
}

// Client retorna el puntero excelize.File.
func (r *Service) Client() *excelize.File { return r.client }

// Address retorna la ruta del archivo.
func (r *Service) Address() string { return r.address }

// Close guarda y cierra el archivo.
func (r *Service) Close() error {
	if r.client == nil {
		return nil
	}
	if err := r.client.Save(); err != nil {
		return err
	}
	return r.client.Close()
}

// Export escribe `data` a la hoja configurada.
func (r *Service) Export(data any) error {
	if r.client == nil {
		return errors.New("client not connected")
	}
	if data == nil {
		return fmt.Errorf("export: data is nil")
	}
	return exportToFile(r.client, r.config, data)
}

// ExportToWriter escribe a un writer (útil HTTP/S3). Crea un file temporal in-memory y vuelca.
func (r *Service) ExportToWriter(data any, w io.Writer) error {
	if data == nil {
		return fmt.Errorf("export: data is nil")
	}
	f := excelize.NewFile()
	defer f.Close()
	if err := exportToFile(f, r.config, data); err != nil {
		return err
	}
	_, err := f.WriteTo(w)
	return err
}

// Import lee la hoja y mapea a dest (puntero a slice de struct o map).
func (r *Service) Import(dest any) error {
	if r.client == nil {
		return errors.New("client not connected")
	}
	if dest == nil {
		return fmt.Errorf("import: dest is nil pointer")
	}
	rows, err := r.client.GetRows(r.config.GetSheet())
	if err != nil {
		return fmt.Errorf("read rows: %w", err)
	}
	if len(rows) == 0 {
		return fmt.Errorf("sheet is empty")
	}
	return mapRowsToDest(r.config, rows, dest)
}

// createWorkbookIfNotExists crea el archivo si no existe.
func (r *Service) createWorkbookIfNotExists(config ConfigPort) error {
	if _, err := os.Stat(config.GetFilePath()); err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("stat: %w", err)
		}
		if dir := filepath.Dir(config.GetFilePath()); dir != "." && dir != "" {
			if mkErr := os.MkdirAll(dir, 0o755); mkErr != nil {
				return fmt.Errorf("mkdir %q: %w", dir, mkErr)
			}
		}
		f := excelize.NewFile()
		defer f.Close()
		// Si la hoja pedida no es "Sheet1", renombramos la default para evitar hojas sobrantes
		if config.GetSheet() != "Sheet1" {
			if err := f.SetSheetName("Sheet1", config.GetSheet()); err != nil {
				return fmt.Errorf("rename default sheet: %w", err)
			}
		}
		if idx, _ := f.GetSheetIndex(config.GetSheet()); idx >= 0 {
			f.SetActiveSheet(idx)
		}
		if err := f.SaveAs(config.GetFilePath()); err != nil {
			return fmt.Errorf("save new workbook: %w", err)
		}
	}
	return nil
}

func exportToFile(f *excelize.File, cfg ConfigPort, data any) error {
	// crear/seleccionar hoja
	idx, err := f.GetSheetIndex(cfg.GetSheet())
	if err != nil || idx < 0 {
		idx, err = f.NewSheet(cfg.GetSheet())
		if err != nil {
			return fmt.Errorf("new sheet %q: %w", cfg.GetSheet(), err)
		}
	}
	f.SetActiveSheet(idx)

	headers, rows, dateCols, err := buildRows(data)
	if err != nil {
		return err
	}

	rowOffset := 1
	if cfg.GetWriteHeader() {
		for c, h := range headers {
			cell, _ := excelize.CoordinatesToCellName(c+1, rowOffset)
			if err := f.SetCellValue(cfg.GetSheet(), cell, h); err != nil {
				return fmt.Errorf("set header %s: %w", cell, err)
			}
		}
		rowOffset++
	}
	for r, row := range rows {
		for c, v := range row {
			cell, _ := excelize.CoordinatesToCellName(c+1, r+rowOffset)
			if err := f.SetCellValue(cfg.GetSheet(), cell, v); err != nil {
				return fmt.Errorf("set cell %s: %w", cell, err)
			}
		}
	}
	if widths := cfg.GetColumnWidths(); len(widths) > 0 {
		for i, w := range widths {
			col, _ := excelize.ColumnNumberToName(i + 1)
			if err := f.SetColWidth(cfg.GetSheet(), col, col, w); err != nil {
				return fmt.Errorf("set col width %s: %w", col, err)
			}
		}
	}
	// Aplicar estilo de fecha solo a columnas con time.Time (detectadas en structs)
	if len(dateCols) > 0 && len(rows) > 0 {
		excelFmt := goDateToExcelNumFmt(cfg.GetDateFormat())
		if excelFmt == "" {
			excelFmt = "yyyy-mm-dd"
		}
		if style, err := f.NewStyle(&excelize.Style{CustomNumFmt: &excelFmt}); err == nil {
			// rango de filas escritas (incluye header si existe)
			startRow := 1
			endRow := rowOffset + len(rows) - 1
			for c := range dateCols {
				colName, _ := excelize.ColumnNumberToName(c + 1)
				start := fmt.Sprintf("%s%d", colName, startRow)
				end := fmt.Sprintf("%s%d", colName, endRow)
				_ = f.SetCellStyle(cfg.GetSheet(), start, end, style)
			}
		}
	}
	return nil
}

func buildRows(data any) ([]string, [][]any, map[int]bool, error) {
	v := reflect.ValueOf(data)
	if v.Kind() != reflect.Slice && v.Kind() != reflect.Array {
		return nil, nil, nil, fmt.Errorf("data must be slice or array")
	}
	if v.Len() == 0 {
		return nil, nil, nil, fmt.Errorf("data slice is empty")
	}
	el := v.Index(0)
	if el.Kind() == reflect.Ptr {
		el = el.Elem()
	}
	switch el.Kind() {
	case reflect.Struct:
		return rowsFromStructs(v)
	case reflect.Map:
		headers, rows, err := rowsFromMaps(v)
		return headers, rows, map[int]bool{}, err
	default:
		return nil, nil, nil, fmt.Errorf("unsupported element kind: %s", el.Kind())
	}
}

func rowsFromStructs(v reflect.Value) ([]string, [][]any, map[int]bool, error) {
	t := v.Index(0).Type()
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	nf := t.NumField()
	idx := make([]int, 0, nf)
	headers := make([]string, 0, nf)
	dateCols := make(map[int]bool)
	for i := 0; i < nf; i++ {
		f := t.Field(i)
		if f.PkgPath != "" {
			continue
		}
		if tag := f.Tag.Get("excel"); tag == "-" {
			continue
		}
		name := f.Tag.Get("excel")
		if name == "" {
			name = f.Tag.Get("json")
		}
		if name == "" || name == "-" {
			name = f.Name
		}
		headers = append(headers, name)
		idx = append(idx, i)
		// marca columna si el tipo base es time.Time (para aplicar estilo)
		ft := f.Type
		if ft.Kind() == reflect.Ptr {
			ft = ft.Elem()
		}
		if ft == reflect.TypeOf(time.Time{}) {
			dateCols[len(headers)-1] = true
		}
	}
	rows := make([][]any, v.Len())
	for r := 0; r < v.Len(); r++ {
		row := make([]any, len(idx))
		item := v.Index(r)
		if item.Kind() == reflect.Ptr {
			item = item.Elem()
		}
		for c, fi := range idx {
			row[c] = toCellValue(item.Field(fi))
		}
		rows[r] = row
	}
	return headers, rows, dateCols, nil
}

func rowsFromMaps(v reflect.Value) ([]string, [][]any, error) {
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
	rows := make([][]any, v.Len())
	for r := 0; r < v.Len(); r++ {
		row := make([]any, len(headers))
		it := v.Index(r)
		if it.Kind() == reflect.Ptr {
			it = it.Elem()
		}
		for c, hk := range headers {
			val := it.MapIndex(reflect.ValueOf(hk))
			row[c] = toCellValue(val)
		}
		rows[r] = row
	}
	return headers, rows, nil
}

// goDateToExcelNumFmt convierte un layout Go (p.ej. "2006-01-02")
// a un mask de Excel (p.ej. "yyyy-mm-dd"). Fallback a "yyyy-mm-dd".
func goDateToExcelNumFmt(goFmt string) string {
	if goFmt == "" {
		return "yyyy-mm-dd"
	}
	// Traducción básica de tokens más comunes.
	repl := []struct{ goTok, xlTok string }{
		{"2006", "yyyy"},
		{"06", "yy"},
		{"01", "mm"},
		{"1", "m"},
		{"02", "dd"},
		{"2", "d"},
		{"15", "hh"}, // 24h
		{"03", "hh"}, // 12h
		{"04", "mm"}, // minutos
		{"05", "ss"},
		{"PM", "AM/PM"},
		{"pm", "am/pm"},
	}
	out := goFmt
	for _, r := range repl {
		out = strings.ReplaceAll(out, r.goTok, r.xlTok)
	}
	return out
}

func toCellValue(v reflect.Value) any {
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
		return v.Int()
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return v.Uint()
	case reflect.Float32, reflect.Float64:
		return v.Float()
	case reflect.Bool:
		return v.Bool()
	case reflect.Struct:
		// Escribir time.Time nativo para que Excel lo interprete como fecha
		if t, ok := v.Interface().(time.Time); ok {
			return t
		}
		// shopspring/decimal
		if d, ok := v.Interface().(decimal.Decimal); ok {
			return d.String()
		}
		// sql.Null*
		switch nv := v.Interface().(type) {
		case sql.NullString:
			if nv.Valid {
				return nv.String
			}
			return ""
		case sql.NullInt64:
			if nv.Valid {
				return nv.Int64
			}
			return ""
		case sql.NullFloat64:
			if nv.Valid {
				return nv.Float64
			}
			return ""
		case sql.NullBool:
			if nv.Valid {
				return nv.Bool
			}
			return ""
		case sql.NullTime:
			if nv.Valid {
				return nv.Time
			}
			return ""
		}
	}
	return fmt.Sprintf("%v", v.Interface())
}

func mapRowsToDest(cfg ConfigPort, rows [][]string, dest any) error {
	if len(rows) == 0 {
		return errors.New("no rows")
	}
	v := reflect.ValueOf(dest)
	if v.Kind() != reflect.Ptr || v.IsNil() {
		return errors.New("dest must be non-nil pointer to slice")
	}
	slice := v.Elem()
	if slice.Kind() != reflect.Slice {
		return errors.New("dest must be pointer to slice")
	}
	headers := rows[0]
	dataRows := rows[1:]
	elType := slice.Type().Elem()
	isPtr := false
	if elType.Kind() == reflect.Ptr {
		isPtr = true
		elType = elType.Elem()
	}
	switch elType.Kind() {
	case reflect.Struct:
		fieldMap := buildFieldMap(elType)
		for _, r := range dataRows {
			el := reflect.New(elType).Elem()
			for i, h := range headers {
				if fi, ok := fieldMap[strings.ToLower(h)]; ok {
					if i < len(r) {
						if err := setFieldFromString(cfg, el.Field(fi), r[i]); err != nil {
							return err
						}
					}
				}
			}
			if isPtr {
				slice = reflect.Append(slice, el.Addr())
			} else {
				slice = reflect.Append(slice, el)
			}
		}
	case reflect.Map:
		if elType.Key().Kind() != reflect.String {
			return errors.New("map key must be string")
		}
		for _, r := range dataRows {
			m := reflect.MakeMapWithSize(elType, len(headers))
			for i, h := range headers {
				if i < len(r) {
					m.SetMapIndex(reflect.ValueOf(h), reflect.ValueOf(r[i]))
				}
			}
			slice = reflect.Append(slice, m)
		}
	default:
		return fmt.Errorf("unsupported slice element kind: %s", elType.Kind())
	}
	v.Elem().Set(slice)
	return nil
}

func buildFieldMap(t reflect.Type) map[string]int {
	m := make(map[string]int)
	n := t.NumField()
	for i := 0; i < n; i++ {
		f := t.Field(i)
		if f.PkgPath != "" {
			continue
		}
		if tag := f.Tag.Get("excel"); tag == "-" {
			continue
		}
		name := f.Tag.Get("excel")
		if name == "" {
			name = f.Tag.Get("json")
		}
		if name == "" || name == "-" {
			name = f.Name
		}
		m[strings.ToLower(name)] = i
	}
	return m
}

func setFieldFromString(cfg ConfigPort, f reflect.Value, val string) error {
	if !f.CanSet() {
		return nil
	}
	if f.Kind() == reflect.Ptr {
		if val == "" {
			return nil
		}
		ptr := reflect.New(f.Type().Elem())
		if err := setFieldFromString(cfg, ptr.Elem(), val); err != nil {
			return err
		}
		f.Set(ptr)
		return nil
	}
	switch f.Kind() {
	case reflect.String:
		f.SetString(val)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		i, err := strconv.ParseInt(val, 10, 64)
		if err != nil {
			return err
		}
		f.SetInt(i)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		u, err := strconv.ParseUint(val, 10, 64)
		if err != nil {
			return err
		}
		f.SetUint(u)
	case reflect.Float32, reflect.Float64:
		fl, err := strconv.ParseFloat(val, 64)
		if err != nil {
			return err
		}
		f.SetFloat(fl)
	case reflect.Bool:
		b, err := strconv.ParseBool(strings.ToLower(val))
		if err != nil {
			return err
		}
		f.SetBool(b)
	case reflect.Struct:
		if f.Type() == reflect.TypeOf(time.Time{}) {
			df := "2006-01-02"
			if cfg != nil && cfg.GetDateFormat() != "" {
				df = cfg.GetDateFormat()
			}
			t, err := time.Parse(df, val)
			if err != nil {
				return err
			}
			f.Set(reflect.ValueOf(t))
			return nil
		}
		// shopspring/decimal.Decimal
		if f.Type() == reflect.TypeOf(decimal.Decimal{}) {
			d, err := decimal.NewFromString(val)
			if err != nil {
				return err
			}
			f.Set(reflect.ValueOf(d))
			return nil
		}
		// sql.Null*
		switch f.Addr().Interface().(type) {
		case *sql.NullString:
			if val == "" {
				f.Set(reflect.ValueOf(sql.NullString{Valid: false}))
			} else {
				f.Set(reflect.ValueOf(sql.NullString{String: val, Valid: true}))
			}
			return nil
		case *sql.NullInt64:
			if val == "" {
				f.Set(reflect.ValueOf(sql.NullInt64{Valid: false}))
				return nil
			}
			i, err := strconv.ParseInt(val, 10, 64)
			if err != nil {
				return err
			}
			f.Set(reflect.ValueOf(sql.NullInt64{Int64: i, Valid: true}))
			return nil
		case *sql.NullFloat64:
			if val == "" {
				f.Set(reflect.ValueOf(sql.NullFloat64{Valid: false}))
				return nil
			}
			fl, err := strconv.ParseFloat(val, 64)
			if err != nil {
				return err
			}
			f.Set(reflect.ValueOf(sql.NullFloat64{Float64: fl, Valid: true}))
			return nil
		case *sql.NullBool:
			if val == "" {
				f.Set(reflect.ValueOf(sql.NullBool{Valid: false}))
				return nil
			}
			b, err := strconv.ParseBool(strings.ToLower(val))
			if err != nil {
				return err
			}
			f.Set(reflect.ValueOf(sql.NullBool{Bool: b, Valid: true}))
			return nil
		case *sql.NullTime:
			df := "2006-01-02"
			if cfg != nil && cfg.GetDateFormat() != "" {
				df = cfg.GetDateFormat()
			}
			if val == "" {
				f.Set(reflect.ValueOf(sql.NullTime{Valid: false}))
				return nil
			}
			t, err := time.Parse(df, val)
			if err != nil {
				return err
			}
			f.Set(reflect.ValueOf(sql.NullTime{Time: t, Valid: true}))
			return nil
		}
		return fmt.Errorf("unsupported struct type: %s", f.Type())
	default:
		return fmt.Errorf("unsupported kind: %s", f.Kind())
	}
	return nil
}
