package pkggin

import (
	"os"

	"github.com/gin-gonic/gin"
)

func Bootstrap(port, apiVersion string, isTest bool) (*Server, error) {
	if gin.Mode() == gin.TestMode {
		return newTestServer()
	}

	if port == "" {
		port = os.Getenv("HTTP_SERVER_PORT")
	}
	if apiVersion == "" {
		apiVersion = os.Getenv("API_VERSION")
	}

	config := newConfig(
		port,
		apiVersion,
	)

	if err := config.Validate(); err != nil {
		return nil, err
	}

	Server, err := newServer(config)
	if err != nil {
		return nil, err
	}

	return Server, nil
}
