package pkgswagger

import (
	"os"
	"strings"
)

func Bootstrap() (Service, error) {
	config := newConfig(
		os.Getenv("SWAGGER_TITLE"),
		os.Getenv("SWAGGER_DESCRIPTION"),
		os.Getenv("SWAGGER_VERSION"),
		os.Getenv("SWAGGER_HOST"),
		os.Getenv("SWAGGER_BASE_PATH"),
		strings.Split(os.Getenv("SWAGGER_SCHEMES"), ","),
		os.Getenv("SWAGGER_ENABLED") == "true",
	)

	if err := config.Validate(); err != nil {
		return nil, err
	}

	return newService(config)
}
