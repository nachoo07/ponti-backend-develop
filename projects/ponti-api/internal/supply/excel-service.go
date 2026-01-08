package supply

import (
	"bytes"
	"context"
	"io"

	types "github.com/alphacodinggroup/ponti-backend/pkg/types"
	supplyExcel "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/supply/excel"
	"github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/supply/usecases/domain"
)

type XLSXEnginePort interface {
	ExportToWriter(rows any, w io.Writer) error
	Close() error
}

type ExcelExporter struct {
	eng XLSXEnginePort
}

func NewExcelExporter(eng XLSXEnginePort) *ExcelExporter {
	return &ExcelExporter{eng: eng}
}

func (e *ExcelExporter) Export(ctx context.Context, items []domain.Supply) ([]byte, error) {
	_ = ctx

	if len(items) == 0 {
		return nil, types.NewError(types.ErrNotFound, "there is no data to export", nil)
	}

	rows := supplyExcel.BuildDTO(items)
	var buf bytes.Buffer
	if err := e.eng.ExportToWriter(rows, &buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (e *ExcelExporter) Close() error { return e.eng.Close() }
