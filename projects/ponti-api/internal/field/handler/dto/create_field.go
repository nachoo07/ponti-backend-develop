package dto

type CreateFieldRequest struct {
    Field
}

type CreateFieldResponse struct {
    Message string `json:"message"`
    ID      int64  `json:"id"`
}
