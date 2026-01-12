package dto

// CreateLot is the DTO for creating a Lot.
type CreateLot struct {
	Lot
}

// CreateLotResponse is the response DTO for CreateLot.
type CreateLotResponse struct {
	Message string `json:"message"`
	ID      int64  `json:"id"`
}
