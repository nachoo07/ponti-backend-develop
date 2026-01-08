package pkgtypes

import (
	"net/http"
	"strconv"
)

// PageInfo contiene la información de paginación de la respuesta JSON.
type PageInfo struct {
	PerPage int   `json:"per_page"`
	Page    int   `json:"page"`
	MaxPage int   `json:"max_page"`
	Total   int64 `json:"total"`
}

// Input representa los datos de entrada para una consulta paginada.
type Input struct {
	Page     uint
	PageSize uint
}

// NewPageInfo crea un PageInfo válido asegurando límites mínimos.
func NewPageInfo(page, perPage int, total int64) PageInfo {
	if perPage < 1 {
		perPage = 1
	}
	if page < 1 {
		page = 1
	}
	maxPage := int((total + int64(perPage) - 1) / int64(perPage))
	if maxPage < 1 {
		maxPage = 1
	}
	return PageInfo{
		PerPage: perPage,
		Page:    page,
		MaxPage: maxPage,
		Total:   total,
	}
}

// GetPaginationFromRequest lee page/per_page/page_size de la querystring.
// Devuelve valores por defecto (1,10) si no están o son inválidos.
func GetPaginationFromRequest(r *http.Request) (page, perPage int) {
	page = 1
	perPage = 10

	if val := r.URL.Query().Get("page"); val != "" {
		if v, err := strconv.Atoi(val); err == nil && v > 0 {
			page = v
		}
	}
	if val := r.URL.Query().Get("per_page"); val != "" {
		if v, err := strconv.Atoi(val); err == nil && v > 0 {
			perPage = v
		}
	} else if val := r.URL.Query().Get("page_size"); val != "" {
		if v, err := strconv.Atoi(val); err == nil && v > 0 {
			perPage = v
		}
	}

	return
}

// NewInput construye un Input para los use-cases/repositorios a partir de la request.
func NewInput(r *http.Request) Input {
	page, perPage := GetPaginationFromRequest(r)
	return Input{
		Page:     uint(page),
		PageSize: uint(perPage),
	}
}
