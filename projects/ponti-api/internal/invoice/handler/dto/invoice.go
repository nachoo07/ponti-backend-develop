package dto

import (
	"time"

	domain "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/invoice/usecases/domain"
)

type InvoiceResponse struct {
	ID          int64     `json:"id"`
	WorkOrderID int64     `json:"work_order_id"`
	Number      string    `json:"number"`
	Company     string    `json:"company"`
	Date        time.Time `json:"date"`
	Status      string    `json:"status"`

	CreateAt  time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func FromDomain(d *domain.Invoice) InvoiceResponse {
	return InvoiceResponse{
		ID:          d.ID,
		WorkOrderID: d.WorkOrderID,
		Number:      d.Number,
		Company:     d.Company,
		Date:        d.Date,
		Status:      d.Status,
		CreateAt:    d.CreatedAt,
		UpdatedAt:   d.UpdatedAt,
	}
}
