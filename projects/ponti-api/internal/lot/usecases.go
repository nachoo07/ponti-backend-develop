// Package lot contiene los casos de uso para la entidad Lot.
package lot

import (
	// standard library
	"context"

	// third-party
	"github.com/shopspring/decimal"

	// project

	types "github.com/alphacodinggroup/ponti-backend/pkg/types"
	domain "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/lot/usecases/domain"
)

type RepositoryPort interface {
	CreateLot(context.Context, *domain.Lot) (int64, error)
	ListLotsByField(context.Context, int64) ([]domain.Lot, error)
	ListLotsByProject(context.Context, int64) ([]domain.Lot, error)
	ListLotsByProjectAndField(context.Context, int64, int64) ([]domain.Lot, error)
	ListLotsByProjectFieldAndCrop(context.Context, int64, int64, int64, string) ([]domain.Lot, error)
	GetLot(context.Context, int64) (*domain.Lot, error)
	UpdateLot(context.Context, *domain.Lot) error
	DeleteLot(context.Context, int64) error
	GetMetrics(context.Context, int64, int64, int64) (*domain.LotMetrics, error)
	ListLots(context.Context, int64, int64, int64, int, int) ([]domain.LotTable, int, decimal.Decimal, decimal.Decimal, error)
	UpdateLotTons(context.Context, int64, decimal.Decimal) error
}

type ExporterAdapterPort interface {
	Export(ctx context.Context, items []domain.LotTable) ([]byte, error)
	Close() error
}

type UseCases struct {
	repo  RepositoryPort
	excel ExporterAdapterPort
}

func NewUseCases(repo RepositoryPort, excel ExporterAdapterPort) *UseCases {
	return &UseCases{repo: repo, excel: excel}
}

func (u *UseCases) CreateLot(ctx context.Context, l *domain.Lot) (int64, error) {
	return u.repo.CreateLot(ctx, l)
}

func (u *UseCases) ListLotsByField(ctx context.Context, fieldID int64) ([]domain.Lot, error) {
	return u.repo.ListLotsByField(ctx, fieldID)
}

func (u *UseCases) GetLot(ctx context.Context, id int64) (*domain.Lot, error) {
	return u.repo.GetLot(ctx, id)
}

func (u *UseCases) UpdateLot(ctx context.Context, l *domain.Lot) error {
	return u.repo.UpdateLot(ctx, l)
}

func (u *UseCases) UpdateLotTons(ctx context.Context, id int64, tons decimal.Decimal) error {
	return u.repo.UpdateLotTons(ctx, id, tons)
}

func (u *UseCases) DeleteLot(ctx context.Context, id int64) error {
	return u.repo.DeleteLot(ctx, id)
}

func (u *UseCases) ListLotsByProject(ctx context.Context, projectID int64) ([]domain.Lot, error) {
	return u.repo.ListLotsByProject(ctx, projectID)
}

func (u *UseCases) ListLotsByProjectAndField(ctx context.Context, projectID, fieldID int64) ([]domain.Lot, error) {
	return u.repo.ListLotsByProjectAndField(ctx, projectID, fieldID)
}

func (u *UseCases) ListLotsByProjectFieldAndCrop(ctx context.Context, projectID, fieldID, cropID int64, cropType string) ([]domain.Lot, error) {
	return u.repo.ListLotsByProjectFieldAndCrop(ctx, projectID, fieldID, cropID, cropType)
}

func (u *UseCases) GetMetrics(
	ctx context.Context,
	projectID, fieldID, cropID int64,
) (*domain.LotMetrics, error) {
	return u.repo.GetMetrics(ctx, projectID, fieldID, cropID)
}

func (u *UseCases) ListLots(
	ctx context.Context,
	projectID, fieldID, cropID int64,
	page, pageSize int,
) ([]domain.LotTable, int, decimal.Decimal, decimal.Decimal, error) {
	return u.repo.ListLots(ctx, projectID, fieldID, cropID, page, pageSize)
}

func (u *UseCases) ExportLots(ctx context.Context, projectID, fieldID, cropID int64, page, pageSize int) ([]byte, error) {
	if u.excel == nil {
		return nil, types.NewError(types.ErrInternal, "exporter not configured", nil)
	}

	items, _, _, _, err := u.ListLots(ctx, projectID, fieldID, cropID, page, pageSize)
	if err != nil {
		return nil, types.NewError(types.ErrInternal, "list lots", err)
	}

	if len(items) == 0 {
		return nil, types.NewError(types.ErrNotFound, "there is no data to export", nil)
	}

	return u.excel.Export(ctx, items)
}
