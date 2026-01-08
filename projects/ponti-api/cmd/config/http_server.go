package config

import (
	"os"
	"strconv"
)

type HTTPServer struct {
	Name string `envconfig:"HTTP_SERVER_NAME" default:"http-server" validate:"required"`
	Host string `envconfig:"HTTP_SERVER_HOST" default:"localhost" validate:"required"`
	Port int    `envconfig:"HTTP_SERVER_PORT" default:"8080" validate:"gte=0"`
}

func (h *HTTPServer) UnmarshalEnv() error {
	// Cloud Run uses PORT environment variable
	// Check for PORT first (Cloud Run), then HTTP_SERVER_PORT
	if portStr := os.Getenv("PORT"); portStr != "" {
		if port, err := strconv.Atoi(portStr); err == nil {
			h.Port = port
		}
	}
	return nil
}
