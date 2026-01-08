package labor

import (
	"context"

	types "github.com/alphacodinggroup/ponti-backend/pkg/types"
	"github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/labor/usecases/domain"
)

type RepositoryPort interface {
	CreateLabor(context.Context, *domain.Labor) (int64, error)
	ListLabor(context.Context, int, int, int64) ([]domain.ListedLabor, int64, error)
	DeleteLabor(context.Context, int64) error
	GetWorkordersByLaborID(ctx context.Context, laborID int64) (int64, error)
	UpdateLabor(context.Context, *domain.Labor) error
	ListLaborCategoriesByTypeID(context.Context, int64) ([]domain.LaborCategory, error)
	ListByWorkorder(context.Context, int64) ([]domain.LaborRawItem, error)
	ListGroupLabor(context.Context, types.Input, int64, int64) ([]domain.LaborListItem, types.PageInfo, error)
	ListAllGroupLabor(context.Context) ([]domain.LaborRawItem, error)
	GetMetrics(context.Context, domain.LaborFilter) (*domain.LaborMetrics, error)
}

type ExporterAdapterPort interface {
	Export(ctx context.Context, items []domain.LaborListItem) ([]byte, error)
	ExportTable(ctx context.Context, items []domain.LaborListItem) ([]byte, error)
	Close() error
}
type UseCases struct {
	repo  RepositoryPort
	excel ExporterAdapterPort
}

func NewUseCases(repo RepositoryPort, excel ExporterAdapterPort) *UseCases {
	return &UseCases{repo: repo, excel: excel}
}

func (u *UseCases) CreateLabor(ctx context.Context, labor *domain.Labor) (int64, error) {
	return u.repo.CreateLabor(ctx, labor)
}

func (u *UseCases) ListLabor(ctx context.Context, page, perPage int, projectID int64) ([]domain.ListedLabor, int64, error) {
	return u.repo.ListLabor(ctx, page, perPage, projectID)
}

func (u *UseCases) DeleteLabor(ctx context.Context, laborID int64) error {
	count, err := u.repo.GetWorkordersByLaborID(ctx, laborID)
	if err != nil {
		return err
	}
	if count > 0 {
		return types.NewError(types.ErrConflict, "labor is being used in a workorder", nil)
	}
	return u.repo.DeleteLabor(ctx, laborID)
}

func (u *UseCases) UpdateLabor(ctx context.Context, labor *domain.Labor) error {
	return u.repo.UpdateLabor(ctx, labor)
}

func (u *UseCases) ListLaborCategoriesByTypeID(ctx context.Context, typeID int64) ([]domain.LaborCategory, error) {
	return u.repo.ListLaborCategoriesByTypeID(ctx, typeID)
}

func (u *UseCases) ListLaborByWorkorder(ctx context.Context, workorderID int64) ([]domain.LaborRawItem, error) {
	return u.repo.ListByWorkorder(ctx, workorderID)
}

func (u *UseCases) ListGroupLaborByWorkorder(ctx context.Context, inp types.Input, projectID int64, fieldID int64) ([]domain.LaborListItem, types.PageInfo, error) {
	rawItems, pageInfo, err := u.repo.ListGroupLabor(ctx, inp, projectID, fieldID)

	// Mapear directamente - NO hacer cálculos manuales (ya vienen de la vista)
	items := make([]domain.LaborListItem, len(rawItems))
	for i, r := range rawItems {
		items[i] = domain.LaborListItem(r)
	}

	return items, pageInfo, err
}

func (u *UseCases) ExportGroupLaborXLSX(ctx context.Context, in types.Input, pid, fid int64) ([]byte, error) {
	if u.excel == nil {
		return nil, types.NewError(types.ErrInternal, "exporter not configured", nil)
	}

	items, _, err := u.ListGroupLaborByWorkorder(ctx, in, pid, fid)
	if err != nil {
		return nil, types.NewError(types.ErrInternal, "list group labor", err)
	}

	if len(items) == 0 {
		return nil, types.NewError(types.ErrNotFound, "there is no data to export", nil)
	}

	return u.excel.Export(ctx, items)
}

func (u *UseCases) ExportAllGroupLabors(ctx context.Context) ([]byte, error) {
	if u.excel == nil {
		return nil, types.NewError(types.ErrInternal, "exporter not configured", nil)
	}

	raw, err := u.repo.ListAllGroupLabor(ctx)
	if err != nil {
		return nil, types.NewError(types.ErrInternal, "list group labor", err)
	}
	items := make([]domain.LaborListItem, len(raw))
	for i, r := range raw {
		items[i] = domain.LaborListItem(r)
	}
	if len(items) == 0 {
		return nil, types.NewError(types.ErrNotFound, "there is no data to export", nil)
	}

	return u.excel.ExportTable(ctx, items)
}

func (u *UseCases) GetMetrics(ctx context.Context, f domain.LaborFilter) (*domain.LaborMetrics, error) {
	return u.repo.GetMetrics(ctx, f)
}
