package dto

import (
	"fmt"
	"time"

	types "github.com/alphacodinggroup/ponti-backend/pkg/types"
	"github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/supply_movement/usecases/domain"
	"github.com/shopspring/decimal"
)

type UpdateSupplyMovementEntryRequest struct {
	Quantity             *decimal.Decimal `json:"quantity"`
	MovementType         *string          `json:"movement_type"`
	MovementDate         *time.Time       `json:"movement_date"`
	ReferenceNumber      *string          `json:"reference_number"`
	ProjectDestinationId *int64           `json:"project_destination_id"`
	SupplyID             *int64           `json:"supply_id"`
	InvestorID           *int64           `json:"investor_id"`
	Provider             *ProviderRequest `json:"provider"`
}

type ProviderRequest struct {
	ID   *int64  `json:"id"`
	Name *string `json:"name"`
}

func (usmr *UpdateSupplyMovementEntryRequest) Validate() error {
	var err error
	if err = validateMovementType(usmr.MovementType); err != nil {
		return err
	}
	if err = validateProjectDestinationId(usmr.ProjectDestinationId, usmr.MovementType); err != nil {
		return err
	}
	if err = validateProvider(usmr.Provider, usmr.MovementType); err != nil {
		return err
	}
	if err = validateSupplyID(usmr.SupplyID); err != nil {
		return err
	}
	if err = validateInvestorId(usmr.InvestorID); err != nil {
		return err
	}

	return err
}

func (usmer *UpdateSupplyMovementEntryRequest) ToDomain(projectId int64, userId *int64, sm *domain.SupplyMovement) *domain.SupplyMovement {
	if usmer.Quantity != nil {
		sm.Quantity = *usmer.Quantity
	}
	if usmer.MovementType != nil {
		sm.MovementType = *usmer.MovementType
	}
	if usmer.MovementDate != nil {
		sm.MovementDate = usmer.MovementDate
	}
	if usmer.ReferenceNumber != nil {
		sm.ReferenceNumber = *usmer.ReferenceNumber
	}
	if usmer.ProjectDestinationId != nil {
		sm.ProjectDestinationId = *usmer.ProjectDestinationId
	}
	if usmer.SupplyID != nil {
		sm.Supply.ID = *usmer.SupplyID
	}
	if usmer.InvestorID != nil {
		sm.Investor.ID = *usmer.InvestorID
	}
	if usmer.Provider != nil {
		sm.Provider.ID = *usmer.Provider.ID
		sm.Provider.Name = *usmer.Provider.Name
	}

	sm.Base.UpdatedBy = userId
	sm.Base.UpdatedAt = time.Now()

	return sm
}

func validateMovementType(movementType *string) error {
	if movementType != nil {
		movementTypeS := *movementType
		if movementTypeS != domain.INTERNAL_MOVEMENT && movementTypeS != domain.OFFICIAL_INVOICE && movementTypeS != domain.STOCK {
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
	}

	return nil
}

func validateSupplyID(supplyId *int64) error {
	if supplyId != nil {
		supplyIdU := *supplyId
		if supplyIdU < 0 {
			return types.NewInvalidIDError("invalid supply_id", nil)
		}
	}

	return nil
}

func validateInvestorId(investorId *int64) error {
	if investorId != nil {
		investorIdU := *investorId
		if investorIdU < 0 {
			return types.NewInvalidIDError("invalid investor_id", nil)
		}
	}

	return nil
}

func validateProjectDestinationId(projectDestinationId *int64, movementType *string) error {

	if projectDestinationId != nil && movementType == nil {
		return types.NewError(types.ErrValidation, "movementType must be  "+domain.INTERNAL_MOVEMENT, nil)
	}

	if projectDestinationId != nil && movementType != nil {
		movementTypeU := *movementType
		projectDestinationIdU := *projectDestinationId
		if movementTypeU == domain.INTERNAL_MOVEMENT && projectDestinationIdU <= 0 {
			return types.NewInvalidIDError("invalid project_destination_id", nil)
		}
	}

	return nil
}

func validateProvider(provider *ProviderRequest, movementType *string) error {
	if provider != nil && movementType == nil {
		return types.NewError(types.ErrValidation, "movementType must be  "+domain.STOCK, nil)
	}

	if provider != nil && movementType != nil {
		movementTypeU := *movementType
		providerU := *provider
		if movementTypeU == domain.STOCK {
			if *providerU.ID <= 0 {
				return types.NewInvalidIDError("invalid provider_id", nil)
			}
			if *providerU.Name == "" {
				return types.NewMissingFieldError("provider_name")
			}
		}
	}

	return nil
}
