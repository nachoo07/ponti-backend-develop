package config

type Deploy struct {
	Environment string `envconfig:"DEPLOY_ENV" validate:"required,oneof=dev stg prod"`
	Platform    string `envconfig:"DEPLOY_PLATFORM" validate:"required,oneof=local docker mix fury aws gcp"`
	Root        string `envconfig:"DEPLOY_PROJECT_ROOT" default:"/app"`
}
