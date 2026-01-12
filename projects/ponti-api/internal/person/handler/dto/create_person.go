package dto

type CreatePerson struct {
	Person
}

// Response
type CreatePersonResponse struct {
	Message  string `json:"message"`
	PersonID string `json:"person_id"`
}
