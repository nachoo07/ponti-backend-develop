package config

type App struct {
	Name       string `envconfig:"APP_NAME" default:"ponti-api" validate:"required"`
	Version    string `envconfig:"APP_VERSION" default:"1.0" validate:"required"`
	MaxRetries int    `envconfig:"APP_MAX_RETRIES" default:"5" validate:"gte=0"`
}
