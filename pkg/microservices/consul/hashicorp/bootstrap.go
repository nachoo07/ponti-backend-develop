package pkgconsul

import (
	"strings"

	"github.com/spf13/viper"
)

// NewConsulInstance crea y devuelve una nueva instancia de Consul
func NewConsulInstance() (Client, error) {
	tagsString := viper.GetString("CONSUL_TAGS")
	tags := strings.Split(tagsString, ",") // Asume que los tags están separados por comas

	config := newConfig(
		viper.GetString("CONSUL_ID"),
		viper.GetString("CONSUL_NAME"),
		viper.GetInt("CONSUL_PORT"),
		viper.GetString("CONSUL_ADDRESS"),
		viper.GetString("CONSUL_SERVICE_NAME"),
		viper.GetString("CONSUL_HEALTH_CHECK"),
		viper.GetString("CONSUL_CHECK_INTERVAL"),
		viper.GetString("CONSUL_CHECK_TIMEOUT"),
		tags,
	)

	if err := config.Validate(); err != nil {
		return nil, err
	}

	return newClient(config)
}
