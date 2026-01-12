package dto

type GetFollowersResponse struct {
	Message   string   `json:"message"`
	Followers []string `json:"followers"`
}
