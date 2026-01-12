package models

import (
	domain "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/lot/usecases/domain"
)

// FromDomainDates convierte las fechas del dominio al modelo de persistencia
func FromDomainDates(lotID int64, dates []domain.LotDates) []LotDates {
	if len(dates) == 0 {
		return nil
	}

	modelDates := make([]LotDates, len(dates))
	for i, date := range dates {
		modelDates[i] = LotDates{
			LotID:       lotID,
			SowingDate:  date.SowingDate,
			HarvestDate: date.HarvestDate,
			Sequence:    date.Sequence,
		}
	}

	return modelDates
}

// ToDomainDates convierte las fechas del modelo de persistencia al dominio
func ToDomainDates(dates []LotDates) []domain.LotDates {
	if len(dates) == 0 {
		return nil
	}

	domainDates := make([]domain.LotDates, len(dates))
	for i, date := range dates {
		domainDates[i] = domain.LotDates{
			SowingDate:  date.SowingDate,
			HarvestDate: date.HarvestDate,
			Sequence:    date.Sequence,
		}
	}

	return domainDates
}
