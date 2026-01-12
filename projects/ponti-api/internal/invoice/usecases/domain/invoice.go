package domain

import (
	"time"

	shareddomain "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/shared/domain"
)

type Invoice struct {
	ID          int64     // ID de cada registro
	WorkOrderID int64     // ID de ordenes de trabajo
	Number      string    // Nro de factura
	Company     string    // Nombre de la empresa
	Date        time.Time // Fecha De Emision
	Status      string    // Estado

	shareddomain.Base
}
