package createsupplymovement

import (
	"fmt"

	types "github.com/alphacodinggroup/ponti-backend/pkg/types"
	investordomain "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/investor/usecases/domain"
	providerdomain "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/provider/usecase/domain"
	shareddomain "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/shared/domain"
	supplydomain "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/supply/usecases/domain"
	"github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/supply_movement/usecases/domain"
	"github.com/shopspring/decimal"

	"time"
)

type CreateSupplyMovementRequestBulk struct {
	SupplyMovements []CreateSupplyMovementEntryRequest `json:"items"`
}

type CreateSupplyMovementEntryRequest struct {
	Quantity             decimal.Decimal `json:"quantity"`
	MovementType         string          `json:"movement_type"`
	MovementDate         *time.Time      `json:"movement_date"`
	Reference            string          `json:"reference_number"`
	ProjectDestinationId int64           `json:"project_destination_id"`
	SupplyID             int64           `json:"supply_id"`
	InvestorID           int64           `json:"investor_id"`
	Provider             ProviderRequest `json:"provider"`
}

type ProviderRequest struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

func (csmr *CreateSupplyMovementEntryRequest) Validate() error {
	var err error
	if err = validateMovementType(csmr.MovementType); err != nil {
		return err
	}
	if err = validateProjectDestinationId(csmr.ProjectDestinationId, csmr.MovementType); err != nil {
		return err
	}
	if err = validateProvider(csmr.Provider, csmr.MovementType); err != nil {
		return err
	}
	if err = validateMovementDate(csmr.MovementDate); err != nil {
		return err
	}
	if err = validateReference(csmr.Reference); err != nil {
		return err
	}
	if err = validateSupplyID(csmr.SupplyID); err != nil {
		return err
	}
	if err = validateInvestorId(csmr.InvestorID); err != nil {
		return err
	}

	return err
}

func (r *CreateSupplyMovementEntryRequest) ToDomain(projectId int64, userId *int64) *domain.SupplyMovement {
	return &domain.SupplyMovement{
		ProjectId:            projectId,
		Quantity:             r.Quantity,
		MovementType:         r.MovementType,
		MovementDate:         r.MovementDate,
		ReferenceNumber:      r.Reference,
		ProjectDestinationId: r.ProjectDestinationId,
		Supply:               &supplydomain.Supply{ID: r.SupplyID},
		Investor:             &investordomain.Investor{ID: r.InvestorID},
		Provider: &providerdomain.Provider{
			ID:   r.Provider.ID,
			Name: r.Provider.Name,
		},
		IsEntry: true,
		Base: shareddomain.Base{
			CreatedBy: userId,
			UpdatedBy: userId,
		},
	}
}

func validateMovementType(movementType string) error {
	if movementType != domain.INTERNAL_MOVEMENT && movementType != domain.OFFICIAL_INVOICE && movementType != domain.STOCK {
		return types.NewError(
			types.ErrValidation,
			fmt.Sprintf(
				"must be a valid type [%s, %s, %s]",
				domain.INTERNAL_MOVEMENT,
				domain.OFFICIAL_INVOICE,
				domain.STOCK,
			),
			nil,
		)
	}
	return nil
}

func validateMovementDate(movementDate *time.Time) error {
	if movementDate == nil {
		return types.NewMissingFieldError("movement_date")
	}
	return nil
}

func validateReference(reference string) error {
	if reference == "" {
		return types.NewMissingFieldError("reference")
	}

	return nil
}

func validateSupplyID(supplyId int64) error {
	if supplyId < 0 {
		return types.NewInvalidIDError("invalid supply_id", nil)
	}

	return nil
}

func validateInvestorId(investorId int64) error {
	if investorId < 0 {
		return types.NewInvalidIDError("invalid investor_id", nil)
	}
	return nil
}

func validateProjectDestinationId(projectDestinationId int64, movementType string) error {
	if movementType == domain.INTERNAL_MOVEMENT && projectDestinationId <= 0 {
		return types.NewInvalidIDError("invalid project_destination_id", nil)
	}
	return nil
}

func validateProvider(provider ProviderRequest, movementType string) error {
	if movementType == domain.STOCK {
		if provider.ID < 0 {
			return types.NewInvalidIDError("invalid provider_id", nil)
		}
		if provider.Name == "" {
			return types.NewMissingFieldError("provider_name")
		}
	}

	return nil
}
