package config

type HTTPServer struct {
	Name string `envconfig:"HTTP_SERVER_NAME" default:"http-server" validate:"required"`
	Host string `envconfig:"HTTP_SERVER_HOST" default:"localhost" validate:"required"`
	Port int    `envconfig:"HTTP_SERVER_PORT" default:"8080" validate:"gte=0"`
}
