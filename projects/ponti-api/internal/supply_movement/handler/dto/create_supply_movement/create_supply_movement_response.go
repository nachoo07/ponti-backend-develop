package createsupplymovement

type CreateSupplyMovementResponse struct{
	SupplyMovementID     int64  `json:"supply_movement_id"`
	IsSaved     bool   `json:"is_saved"`
	ErrorDetail string `json:"error_detail"`
}

type CreateSupplyMovementBulkResponse struct {
	SupplyMovements []CreateSupplyMovementResponse `json:"supply_movements"`
}

func NewSuccessfulCreateSupplyMovementResponse(supplyMovementId int64) CreateSupplyMovementResponse{
	return CreateSupplyMovementResponse{
		SupplyMovementID: supplyMovementId,
		IsSaved: true,
	}
}

func NewErrorCreateSupplyMovementResponse(errorDetail string)  CreateSupplyMovementResponse{
	return CreateSupplyMovementResponse{
		IsSaved: false,
		ErrorDetail: errorDetail,
	}
}