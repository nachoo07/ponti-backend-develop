package config

type API struct {
	Version string `envconfig:"API_VERSION" default:"v1" validate:"required"`
	BaseURL string // BaseURL es la URL base de la API
}

func (c *API) APIVersion() string {
	return c.Version
}

func (c *API) APIBaseURL() string {
	return c.BaseURL
}
