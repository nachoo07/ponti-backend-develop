package pkggorhttp

import (
	"os"
)

func Bootstrap(port, version string) (Server, error) {
	if port == "" {
		port = os.Getenv("HTTP_SERVER_PORT")
	}
	if version == "" {
		version = os.Getenv("API_VERSION")
	}
	config := newConfig(port, version)
	if err := config.Validate(); err != nil {
		return nil, err
	}
	return newServer(config)
}
