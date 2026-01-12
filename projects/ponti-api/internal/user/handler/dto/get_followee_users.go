package dto

// Response
type GetFolloweesResponse struct {
	Message   string   `json:"message"`
	Followees []string `json:"followees"`
}
