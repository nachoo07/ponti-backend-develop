package workorder

import (
	"context"

	"github.com/shopspring/decimal"

	types "github.com/alphacodinggroup/ponti-backend/pkg/types"
	"github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/workorder/usecases/domain"
)

type RepositoryPort interface {
	CreateWorkorder(context.Context, *domain.Workorder) (int64, error)
	GetWorkorderByID(ctx context.Context, id int64) (*domain.Workorder, error)
	GetWorkorderByNumberAndProjectID(ctx context.Context, number string, projectID int64) (*domain.Workorder, error)
	UpdateWorkorderByID(context.Context, *domain.Workorder) error
	DeleteWorkorderByID(context.Context, int64) error
	ListWorkorders(context.Context, domain.WorkorderFilter, types.Input) ([]domain.WorkorderListElement, types.PageInfo, error)
	GetMetrics(context.Context, domain.WorkorderFilter) (*domain.WorkorderMetrics, error)
	GetRawDirectCost(context.Context, int64) (decimal.Decimal, error)
}

type ExporterAdapterPort interface {
	Export(ctx context.Context, items []domain.WorkorderListElement) ([]byte, error)
	Close() error
}

type UseCases struct {
	repo  RepositoryPort
	excel ExporterAdapterPort
}

func NewUseCases(r RepositoryPort, excel ExporterAdapterPort) *UseCases {
	return &UseCases{repo: r, excel: excel}
}

func (u *UseCases) CreateWorkorder(ctx context.Context, o *domain.Workorder) (int64, error) {
	workorder, err := u.repo.GetWorkorderByNumberAndProjectID(ctx, o.Number, o.ProjectID)
	if err != nil {
		return 0, err
	}
	if workorder != nil {
		return 0, types.NewError(types.ErrConflict, "workorder already exists", nil)
	}

	return u.repo.CreateWorkorder(ctx, o)
}

func (u *UseCases) GetWorkorderByID(ctx context.Context, id int64) (*domain.Workorder, error) {
	return u.repo.GetWorkorderByID(ctx, id)
}

func (u *UseCases) DuplicateWorkorder(ctx context.Context, number string) (string, error) {
	return "", nil
}

func (u *UseCases) UpdateWorkorderByID(ctx context.Context, o *domain.Workorder) error {
	return u.repo.UpdateWorkorderByID(ctx, o)
}

func (u *UseCases) DeleteWorkorderByID(ctx context.Context, id int64) error {
	return u.repo.DeleteWorkorderByID(ctx, id)
}

// ListWorkorders delega al repositorio
func (u *UseCases) ListWorkorders(
	ctx context.Context,
	filt domain.WorkorderFilter,
	inp types.Input,
) ([]domain.WorkorderListElement, types.PageInfo, error) {
	return u.repo.ListWorkorders(ctx, filt, inp)
}

// GetMetrics delega al repositorio
func (u *UseCases) GetMetrics(ctx context.Context, f domain.WorkorderFilter) (*domain.WorkorderMetrics, error) {
	return u.repo.GetMetrics(ctx, f)
}

func (u *UseCases) ExportWorkorders(ctx context.Context, filt domain.WorkorderFilter, inp types.Input) ([]byte, error) {
	if u.excel == nil {
		return nil, types.NewError(types.ErrInternal, "exporter not configured", nil)
	}

	items, _, err := u.ListWorkorders(ctx, filt, inp)
	if err != nil {
		return nil, types.NewError(types.ErrInternal, "list Workorders", err)
	}

	if len(items) == 0 {
		return nil, types.NewError(types.ErrNotFound, "there is no data to export", nil)
	}

	return u.excel.Export(ctx, items)
}
