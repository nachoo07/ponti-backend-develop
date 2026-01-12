package pkgafka

import (
	"github.com/spf13/viper"
)

func Bootstrap(brokersKey, groupIDKey string) (Service, error) {
	config := newConfig(
		viper.GetStringSlice(brokersKey),
		viper.GetString(groupIDKey),
	)

	if err := config.Validate(); err != nil {
		return nil, err
	}

	return newService(config)
}
