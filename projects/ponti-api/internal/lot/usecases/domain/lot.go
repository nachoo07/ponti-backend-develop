package domain

import (
	"time"

	cropdom "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/crop/usecases/domain"
	shareddomain "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/shared/domain"
	"github.com/shopspring/decimal"
)

type Lot struct {
	ID            int64
	Name          string
	FieldID       int64
	Hectares      decimal.Decimal
	PreviousCrop  cropdom.Crop
	CurrentCrop   cropdom.Crop
	Variety       string
	SowingDate    *time.Time
	Season        string
	Status        string
	Dates         []LotDates
	Cost          decimal.Decimal // Costo por hectárea
	HarvestedTons decimal.Decimal // Toneladas cosechadas
	Tons          decimal.Decimal // Toneladas cosechadas

	shareddomain.Base // <-- embebe campos de auditoría
}

type LotMetrics struct {
	SeededArea      decimal.Decimal
	HarvestedArea   decimal.Decimal
	YieldTnPerHa    decimal.Decimal
	CostPerHectare  decimal.Decimal
	SuperficieTotal decimal.Decimal
}
