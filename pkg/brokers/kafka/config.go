package pkgafka

import (
	"fmt"
)

type config struct {
	brokers []string
	groupID string
}

// newConfig crea una nueva configuración para Kafka
func newConfig(brokers []string, groupID string) Config {
	return &config{
		brokers: brokers,
		groupID: groupID,
	}
}

// GetBrokers devuelve la lista de brokers de Kafka
func (c *config) GetBrokers() []string {
	return c.brokers
}

// GetGroupID devuelve el ID de grupo para el consumidor de Kafka
func (c *config) GetGroupID() string {
	return c.groupID
}

// Validate verifica que la configuración de Kafka sea válida
func (c *config) Validate() error {
	if len(c.brokers) == 0 {
		return fmt.Errorf("Kafka brokers are not configured")
	}
	if c.groupID == "" {
		return fmt.Errorf("Kafka group ID is not configured")
	}
	return nil
}
