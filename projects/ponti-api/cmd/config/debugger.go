package config

type Debugger struct {
	Port int `envconfig:"DEBUGGER_PORT" default:"2345" validate:"gte=0"`
}
