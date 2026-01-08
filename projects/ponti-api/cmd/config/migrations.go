package config

type Migrations struct {
	Dir string `envconfig:"MIGRATIONS_DIR" default:"file://migrations"`
}
