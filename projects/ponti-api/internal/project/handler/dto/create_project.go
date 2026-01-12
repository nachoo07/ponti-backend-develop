package dto

type CreateProject struct {
	Project
}

type CreateProjectResponse struct {
	Message   string `json:"message"`
	ProjectID int64  `json:"project"`
}
