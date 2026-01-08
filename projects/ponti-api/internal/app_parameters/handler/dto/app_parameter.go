package dto

import (
	domain "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/app_parameters/usecases/domain"
)

type AppParameterResponse struct {
	ID          int64  `json:"id"`
	Key         string `json:"key"`
	Value       string `json:"value"`
	Type        string `json:"type"`
	Category    string `json:"category"`
	Description string `json:"description"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

type CreateAppParameterRequest struct {
	Key         string `json:"key" binding:"required"`
	Value       string `json:"value" binding:"required"`
	Type        string `json:"type" binding:"required,oneof=decimal integer string boolean"`
	Category    string `json:"category" binding:"required,oneof=units calculations business_rules"`
	Description string `json:"description"`
}

type UpdateAppParameterRequest struct {
	Key         string `json:"key" binding:"required"`
	Value       string `json:"value" binding:"required"`
	Type        string `json:"type" binding:"required,oneof=decimal integer string boolean"`
	Category    string `json:"category" binding:"required,oneof=units calculations business_rules"`
	Description string `json:"description"`
}

func FromDomain(param *domain.AppParameter) AppParameterResponse {
	return AppParameterResponse{
		ID:          param.ID,
		Key:         param.Key,
		Value:       param.Value,
		Type:        param.Type,
		Category:    param.Category,
		Description: param.Description,
		CreatedAt:   param.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:   param.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}

func (req *CreateAppParameterRequest) ToDomain() *domain.AppParameter {
	return &domain.AppParameter{
		Key:         req.Key,
		Value:       req.Value,
		Type:        req.Type,
		Category:    req.Category,
		Description: req.Description,
	}
}

func (req *UpdateAppParameterRequest) ToDomain(id int64) *domain.AppParameter {
	return &domain.AppParameter{
		ID:          id,
		Key:         req.Key,
		Value:       req.Value,
		Type:        req.Type,
		Category:    req.Category,
		Description: req.Description,
	}
}
