package pkgrabbit

import (
	"fmt"

	"github.com/spf13/viper"
)

func Bootstrap() (Broker, error) {
	config := newConfig(
		viper.GetString("RABBITMQ_SERVICE_NAME"),
		viper.GetString("RABBITMQ_HOST"),
		viper.GetString("RABBITMQ_USER"),
		viper.GetString("RABBITMQ_PASSWORD"),
		viper.GetString("RABBITMQ_VHOST"),
		viper.GetString("RABBITMQ_QUEUE"),
		viper.GetString("RABBITMQ_EXCHANGE"),
		viper.GetString("RABBITMQ_EXCHANGE_TYPE"),
		viper.GetString("RABBITMQ_ROUTING_KEY"),
		viper.GetInt("RABBITMQ_PORT"),
		viper.GetBool("RABBITMQ_AUTO_ACK"),
		viper.GetBool("RABBITMQ_EXCLUSIVE"),
		viper.GetBool("RABBITMQ_NO_LOCAL"),
		viper.GetBool("RABBITMQ_NO_WAIT"),
	)

	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("configuration validation failed: %w", err)
	}

	return newService(config)
}
