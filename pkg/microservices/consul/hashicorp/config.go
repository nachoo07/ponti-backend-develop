package pkgconsul

import (
	"fmt"
)

// config define la configuración para el cliente de Consul
type config struct {
	ID            string
	Name          string
	Port          int
	Address       string
	Service       string
	HealthCheck   string
	CheckInterval string
	CheckTimeout  string
	Tags          []string
}

func newConfig(id, name string, port int, address, service, healthCheck, checkInterval, checkTimeout string, tags []string) config {
	return config{
		ID:            id,
		Name:          name,
		Port:          port,
		Address:       address,
		Service:       service,
		HealthCheck:   healthCheck,
		CheckInterval: checkInterval,
		CheckTimeout:  checkTimeout,
		Tags:          tags,
	}
}

func (c config) Validate() error {
	if c.ID == "" || c.Name == "" || c.Port == 0 || c.Address == "" || c.HealthCheck == "" || c.Service == "" {
		return fmt.Errorf("incomplete Consul configuration")
	}
	return nil
}
