package config

type Loader interface {
	GetAppConfig() AppConfig
}
