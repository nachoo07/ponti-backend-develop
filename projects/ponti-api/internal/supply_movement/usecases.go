package supply_movement

import (
	"context"
	"fmt"
	"time"

	types "github.com/alphacodinggroup/ponti-backend/pkg/types"
	projdom "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/project/usecases/domain"
	providerdomain "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/provider/usecase/domain"
	stockUseCases "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/stock"
	stockdomain "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/stock/usecases/domain"
	"github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/supply_movement/usecases/domain"
	"github.com/shopspring/decimal"
)

type RepositoryPort interface {
	GetEntriesSupplyMovementsByProjectID(context.Context, int64) ([]*domain.SupplyMovement, error)
	CreateSupplyMovement(context.Context, *domain.SupplyMovement) (int64, error)
	CreateProvider(context.Context, *providerdomain.Provider) (int64, error)
	GetSupplyMovementByID(context.Context, int64) (*domain.SupplyMovement, error)
	UpdateSupplyMovement(context.Context, *domain.SupplyMovement) error
	GetProviders(context.Context) ([]providerdomain.Provider, error)
	DeleteSupplyMovement(context.Context, int64, int64) error
}

type ExporterAdapterPort interface {
	Export(ctx context.Context, items []*domain.SupplyMovement) ([]byte, error)
	Close() error
}

type UseCases struct {
	repo          RepositoryPort
	stockUseCases stockUseCases.UseCasesPort
	excel         ExporterAdapterPort
}

func NewUseCases(
	repo RepositoryPort,
	stockUseCases stockUseCases.UseCasesPort,
	excel ExporterAdapterPort,
) *UseCases {
	return &UseCases{repo: repo, stockUseCases: stockUseCases, excel: excel}
}

func (u *UseCases) CreateSupplyMovement(ctx context.Context, movement *domain.SupplyMovement) (int64, error) {
	stock, isFirst, err := u.stockUseCases.GetLastStockByProjectId(ctx, movement.ProjectId, movement.Supply.ID)
	if err != nil {
		return 0, err
	}
	if isFirst {
		stock = createStockDomainFromSupplyMovement(movement)
		stockId, err := u.stockUseCases.CreateStock(ctx, stock)
		if err != nil {
			return 0, err
		}
		stock.ID = stockId
	}

	if movement.MovementType == domain.INTERNAL_MOVEMENT {
		err := u.handleMovementInternalMovementOut(ctx, movement, *stock)
		if err != nil {
			return 0, err
		}
		// Retornar directamente - handleMovementInternalMovementOut ya creó todos los registros necesarios
		// y ya actualizó los stocks de origen y destino
		return 0, nil
	}
	movement.StockId = stock.ID

	stockDiference := createStockDiference(movement.IsEntry, movement.Quantity)

	stock.RealStockUnits = stock.RealStockUnits.Add(stockDiference)

	err = u.stockUseCases.UpdateRealStockUnits(ctx, stock.ID, stock)
	if err != nil {
		return 0, err
	}

	if movement.Provider.ID == 0 {
		providerID, err := u.repo.CreateProvider(ctx, movement.Provider)
		if err != nil {
			return 0, err
		}
		movement.Provider.ID = providerID
	}

	return u.repo.CreateSupplyMovement(ctx, movement)
}

func (u *UseCases) GetEntriesSupplyMovementsByProjectID(ctx context.Context, projectId int64) ([]*domain.SupplyMovement, error) {
	return u.repo.GetEntriesSupplyMovementsByProjectID(ctx, projectId)
}

func (u *UseCases) UpdateSupplyMovement(ctx context.Context, supplyMovement *domain.SupplyMovement) error {
	return u.repo.UpdateSupplyMovement(ctx, supplyMovement)
}

func (u *UseCases) GetSupplyMovementByID(ctx context.Context, id int64) (*domain.SupplyMovement, error) {
	return u.repo.GetSupplyMovementByID(ctx, id)
}

func (u *UseCases) DeleteSupplyMovement(ctx context.Context, projectId, supplyId int64) error {
	return u.repo.DeleteSupplyMovement(ctx, projectId, supplyId)
}

func (u *UseCases) GetProviders(ctx context.Context) ([]providerdomain.Provider, error) {
	return u.repo.GetProviders(ctx)
}

func createStockDomainFromSupplyMovement(supplyMovement *domain.SupplyMovement) *stockdomain.Stock {
	return &stockdomain.Stock{
		Project: &projdom.Project{
			ID: supplyMovement.ProjectId,
		},
		Supply:       supplyMovement.Supply,
		Investor:     supplyMovement.Investor,
		CloseDate:    nil,
		InitialStock: supplyMovement.Quantity,
		YearPeriod:   int64(time.Now().Year()),
		MonthPeriod:  int64(time.Now().Month()),
		Base:         supplyMovement.Base,
	}
}

func createStockDiference(isEntry bool, quantity decimal.Decimal) decimal.Decimal {
	if isEntry {
		return quantity
	} else {
		return quantity.Neg()
	}
}

func (u *UseCases) handleMovementInternalMovementOut(ctx context.Context, movement *domain.SupplyMovement, stockOrigin stockdomain.Stock) error {
	if stockOrigin.RealStockUnits.LessThan(movement.Quantity) {
		return types.NewError(types.ErrValidation, fmt.Sprintf("La cantidad que desea mover es mayor al stock real: %s", stockOrigin.RealStockUnits.String()), nil)
	}

	// Asegurar que el provider existe antes de crear los registros
	if movement.Provider.ID == 0 {
		providerID, err := u.repo.CreateProvider(ctx, movement.Provider)
		if err != nil {
			return err
		}
		movement.Provider.ID = providerID
	}

	// Actualizar el stock del proyecto origen (restar la cantidad)
	stockOrigin.RealStockUnits = stockOrigin.RealStockUnits.Sub(movement.Quantity)
	err := u.stockUseCases.UpdateRealStockUnits(ctx, stockOrigin.ID, &stockOrigin)
	if err != nil {
		return fmt.Errorf("error updating origin stock: %w", err)
	}

	// Crear registro de salida con cantidad NEGATIVA para el proyecto origen
	// Esto representa la salida de inversión y aparecerá en la lista de insumos
	movementOut := *movement
	movementOut.Quantity = movement.Quantity.Neg() // ⚡ CANTIDAD NEGATIVA (dinero negativo se calcula automáticamente: price * cantidad_negativa)
	movementOut.MovementType = domain.INTERNAL_MOVEMENT
	movementOut.IsEntry = true           // ✅ IsEntry=true para que aparezca en el listado de insumos
	movementOut.ProjectDestinationId = 0 // Limpiar destino, este es el registro del origen
	movementOut.StockId = stockOrigin.ID // Asociar al stock del proyecto origen

	// Guardar el registro de salida
	_, err = u.repo.CreateSupplyMovement(ctx, &movementOut)
	if err != nil {
		return fmt.Errorf("error creating internal movement out record: %w", err)
	}

	// Crear registro de entrada con cantidad POSITIVA para el proyecto destino
	movementIn := *movement
	movementIn.ProjectId = movement.ProjectDestinationId
	movementIn.MovementType = domain.INTERNAL_MOVEMENT_IN
	movementIn.IsEntry = true
	movementIn.ProjectDestinationId = 0

	// Buscar o crear el stock en el proyecto destino
	stockDest, isFirstDest, err := u.stockUseCases.GetLastStockByProjectId(ctx, movementIn.ProjectId, movement.Supply.ID)
	if err != nil {
		return fmt.Errorf("error getting destination stock: %w", err)
	}
	if isFirstDest {
		stockDest = createStockDomainFromSupplyMovement(&movementIn)
		stockIdDest, err := u.stockUseCases.CreateStock(ctx, stockDest)
		if err != nil {
			return fmt.Errorf("error creating destination stock: %w", err)
		}
		stockDest.ID = stockIdDest
	}

	// Asignar el stock del proyecto destino
	movementIn.StockId = stockDest.ID

	// Actualizar el stock real del proyecto destino
	stockDest.RealStockUnits = stockDest.RealStockUnits.Add(movementIn.Quantity)
	err = u.stockUseCases.UpdateRealStockUnits(ctx, stockDest.ID, stockDest)
	if err != nil {
		return fmt.Errorf("error updating destination stock: %w", err)
	}

	// Crear el movimiento de entrada directamente (sin recursión)
	_, err = u.repo.CreateSupplyMovement(ctx, &movementIn)
	if err != nil {
		return fmt.Errorf("error creating internal movement in record: %w", err)
	}

	// Marcar el movimiento original como salida (para el tracking del stock físico)
	movement.IsEntry = false

	return nil
}

func (u *UseCases) ExportSupplyMovementsByProjectID(ctx context.Context, projectID int64) ([]byte, error) {
	if u.excel == nil {
		return nil, types.NewError(types.ErrInternal, "exporter not configured", nil)
	}

	items, err := u.GetEntriesSupplyMovementsByProjectID(ctx, projectID)
	if err != nil {
		return nil, types.NewError(types.ErrInternal, "list Supply Movements", err)
	}

	if len(items) == 0 {
		return nil, types.NewError(types.ErrNotFound, "there is no data to export", nil)
	}

	return u.excel.Export(ctx, items)
}
