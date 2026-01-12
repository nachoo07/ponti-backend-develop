package wire

import (
	config "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/cmd/config"
)

func ProvideConfigLoader() (config.Loader, error) {
	return config.NewConfigLoader()
}
