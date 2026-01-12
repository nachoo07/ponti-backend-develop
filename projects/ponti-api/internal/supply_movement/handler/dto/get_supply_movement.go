package dto

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/shopspring/decimal"

	"github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/supply_movement/usecases/domain"
)

type GetEntrySupplyMovementsResponse struct {
	Summary                       summary                        `json:"summary"`
	EntrySupplyMovementsResponses []entrySupplyMovementsResponse `json:"entries"`
}

type summary struct {
	TotalKg  decimal.Decimal `json:"total_kg"`
	TotalLt  decimal.Decimal `json:"total_lt"`
	TotalUSD decimal.Decimal `json:"total_usd"`
}

// MarshalJSON aplica redondeo al entero más cercano para todos los campos
func (s summary) MarshalJSON() ([]byte, error) {
	aux := struct {
		TotalKg  string `json:"total_kg"`
		TotalLt  string `json:"total_lt"`
		TotalUSD string `json:"total_usd"`
	}{
		TotalKg:  s.TotalKg.Round(0).String(),  // Redondeo al entero más cercano
		TotalLt:  s.TotalLt.Round(0).String(),  // Redondeo al entero más cercano
		TotalUSD: s.TotalUSD.Round(0).String(), // Redondeo al entero más cercano
	}
	return json.Marshal(aux)
}

type entrySupplyMovementsResponse struct {
	ID              int64           `json:"id"`
	EntryType       string          `json:"entry_type"`
	ReferenceNumber string          `json:"reference_number"`
	EntryDate       time.Time       `json:"entry_date"`
	InvestorName    string          `json:"investor_name"`
	SupplyName      string          `json:"supply_name"`
	Quantity        string          `json:"quantity"`
	Category        string          `json:"category"`
	Type            string          `json:"type"`
	ProviderName    string          `json:"provider_name"`
	PriceUSD        decimal.Decimal `json:"price_usd"`
	TotalUSD        decimal.Decimal `json:"total_usd"`
}

func entrySupplyMovementsResponseFromDomain(dsm *domain.SupplyMovement) entrySupplyMovementsResponse {
	return entrySupplyMovementsResponse{
		ID:              dsm.ID,
		EntryType:       dsm.MovementType,
		ReferenceNumber: dsm.ReferenceNumber,
		EntryDate:       *dsm.MovementDate,
		InvestorName:    dsm.Investor.Name,
		SupplyName:      dsm.Supply.Name,
		Quantity:        fmt.Sprintf("%s %s", dsm.Quantity.String(), dsm.Supply.UnitName),
		Category:        dsm.Supply.CategoryName,
		Type:            dsm.Supply.Type.Name,
		PriceUSD:        dsm.Supply.Price,
		TotalUSD:        dsm.Supply.Price.Mul(dsm.Quantity),
		ProviderName:    dsm.Provider.Name,
	}
}

func NewGetEntrySupplyMovementsResponse(entriesDomain []*domain.SupplyMovement) GetEntrySupplyMovementsResponse {
	var totalKg decimal.Decimal
	var totalLt decimal.Decimal
	var totalUSD decimal.Decimal
	var entrySupplyMovementsResponses []entrySupplyMovementsResponse

	for i, supplyMovement := range entriesDomain {
		entrySupplyMovementsResponses = append(
			entrySupplyMovementsResponses,
			entrySupplyMovementsResponseFromDomain(supplyMovement),
		)

		// Usar unit_id directamente en lugar de buscar strings en el nombre
		// unit_id = 1 → Lt (litros)
		// unit_id = 2 → Kg (kilos)
		if supplyMovement.Supply.UnitID == 2 {
			totalKg = totalKg.Add(supplyMovement.Quantity)
		} else if supplyMovement.Supply.UnitID == 1 {
			totalLt = totalLt.Add(supplyMovement.Quantity)
		}
		totalUSD = totalUSD.Add(entrySupplyMovementsResponses[i].TotalUSD)
	}

	summary := summary{
		TotalKg:  totalKg,
		TotalLt:  totalLt,
		TotalUSD: totalUSD,
	}

	return GetEntrySupplyMovementsResponse{
		Summary:                       summary,
		EntrySupplyMovementsResponses: entrySupplyMovementsResponses,
	}
}

func isKG(unitName string) bool {
	return strings.Contains(strings.ToLower(unitName), "kg")
}

func isLt(unitName string) bool {
	return strings.Contains(strings.ToLower(unitName), "lt")
}
