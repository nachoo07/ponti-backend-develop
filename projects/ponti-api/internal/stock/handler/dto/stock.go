package dto

import (
	"encoding/json"
	"strings"
	"time"

	"github.com/shopspring/decimal"

	domain "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/stock/usecases/domain"
)

type GetStocksResponse struct {
	Stocks         []GetStockSummary `json:"items"`
	NetTotalUSD    decimal.Decimal   `json:"net_total_usd"`
	TotalLiters    decimal.Decimal   `json:"total_liters"`
	TotalKilograms decimal.Decimal   `json:"total_kilograms"`
}

// MarshalJSON aplica redondeo: Total U$/Neto, Total insumos (kg), Total insumos (lts) al entero más próximo
func (r GetStocksResponse) MarshalJSON() ([]byte, error) {
	aux := struct {
		Stocks         []GetStockSummary `json:"items"`
		NetTotalUSD    string            `json:"net_total_usd"`
		TotalLiters    string            `json:"total_liters"`
		TotalKilograms string            `json:"total_kilograms"`
	}{
		Stocks:         r.Stocks,
		NetTotalUSD:    r.NetTotalUSD.StringFixed(0),    // Total u$s: entero más próximo
		TotalLiters:    r.TotalLiters.StringFixed(0),    // Total insumos (lts): entero más próximo
		TotalKilograms: r.TotalKilograms.StringFixed(0), // Total insumos (kg): entero más próximo
	}
	return json.Marshal(aux)
}

type GetStockSummary struct {
	ID              int64           `json:"id"`
	SupplyName      string          `json:"supply_name"`
	InvestorName    string          `json:"investor_name"`
	StockUnits      decimal.Decimal `json:"stock_units"`
	RealStockUnits  decimal.Decimal `json:"real_stock_units"`
	StockDifference decimal.Decimal `json:"stock_difference"`
	TotalUSD        decimal.Decimal `json:"total_usd"`
	ClassType       string          `json:"class_type"`
	CloseDate       *time.Time      `json:"close_date"`
	SupplyUnitId    int64           `json:"supply_unit_id"`
	SupplyUnitPrice decimal.Decimal `json:"supply_unit_price"`
	EntryStock      decimal.Decimal `json:"entry_stock"`
	OutStock        decimal.Decimal `json:"out_stock"`
	Consumed        decimal.Decimal `json:"consumed"`
}

// MarshalJSON aplica redondeo: Precio u: 2 dec, Total u$s: 2 dec
func (s GetStockSummary) MarshalJSON() ([]byte, error) {
	aux := struct {
		ID              int64      `json:"id"`
		SupplyName      string     `json:"supply_name"`
		InvestorName    string     `json:"investor_name"`
		StockUnits      string     `json:"stock_units"`
		RealStockUnits  string     `json:"real_stock_units"`
		StockDifference string     `json:"stock_difference"`
		TotalUSD        string     `json:"total_usd"`
		ClassType       string     `json:"class_type"`
		CloseDate       *time.Time `json:"close_date"`
		SupplyUnitId    int64      `json:"supply_unit_id"`
		SupplyUnitPrice string     `json:"supply_unit_price"`
		EntryStock      string     `json:"entry_stock"`
		OutStock        string     `json:"out_stock"`
		Consumed        string     `json:"consumed"`
	}{
		ID:              s.ID,
		SupplyName:      s.SupplyName,
		InvestorName:    s.InvestorName,
		StockUnits:      s.StockUnits.StringFixed(2),
		RealStockUnits:  s.RealStockUnits.StringFixed(2),
		StockDifference: s.StockDifference.StringFixed(2),
		TotalUSD:        s.TotalUSD.StringFixed(2), // Total u$s: 2 decimales
		ClassType:       s.ClassType,
		CloseDate:       s.CloseDate,
		SupplyUnitId:    s.SupplyUnitId,
		SupplyUnitPrice: s.SupplyUnitPrice.StringFixed(2), // Precio u: 2 decimales
		EntryStock:      s.EntryStock.StringFixed(2),
		OutStock:        s.OutStock.StringFixed(2),
		Consumed:        s.Consumed.StringFixed(2),
	}
	return json.Marshal(aux)
}

// FromDomain maps domain.Stock to GetStock DTO
func FromDomain(s *domain.Stock) *GetStockSummary {
	return &GetStockSummary{
		ID:              s.ID,
		InvestorName:    s.Investor.Name,
		SupplyName:      s.Supply.Name,
		StockUnits:      s.GetStockUnits(),
		RealStockUnits:  s.RealStockUnits,
		TotalUSD:        s.GetTotalUSD(),
		StockDifference: s.GetStockDifference(),
		CloseDate:       s.CloseDate,
		ClassType:       s.Supply.CategoryName, // FIX: usar CategoryName (Herbicidas, Coadyuvantes) en lugar de Type.Name (Agroquímicos)
		SupplyUnitId:    s.Supply.UnitID,
		SupplyUnitPrice: s.Supply.Price,
		EntryStock:      s.GetEntryStock(),
		OutStock:        s.GetOutStock(),
		Consumed:        s.Consumed,
	}
}

func NewGetStocksListed(stocks []*domain.Stock) GetStocksResponse {
	var netTotalUSD decimal.Decimal
	var totalKilograms decimal.Decimal
	var totalLiters decimal.Decimal
	listedStocks := make([]GetStockSummary, len(stocks))

	for i, stock := range stocks {
		listedStocks[i] = *FromDomain(stock)
		netTotalUSD = netTotalUSD.Add(stock.GetTotalUSD())
		if isKG(stock.GetSupplyUnitName()) {
			totalKilograms = totalKilograms.Add(stock.GetStockUnits())
		} else if isLt(stock.GetSupplyUnitName()) {
			totalLiters = totalLiters.Add(stock.GetStockUnits())
		}
	}

	return GetStocksResponse{
		Stocks:         listedStocks,
		NetTotalUSD:    netTotalUSD,
		TotalLiters:    totalLiters,
		TotalKilograms: totalKilograms,
	}

}

func isKG(unitName string) bool {
	return strings.Contains(strings.ToLower(unitName), "kg")
}

func isLt(unitName string) bool {
	return strings.Contains(strings.ToLower(unitName), "lt")
}
