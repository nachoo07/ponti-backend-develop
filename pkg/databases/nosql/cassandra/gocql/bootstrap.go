package pkgcassandra

import (
	"os"
	"strings"
)

func Bootstrap() (Repository, error) {
	// Obtener la cadena de hosts y convertirla en un slice (suponiendo que estén separados por comas)
	hostsStr := os.Getenv("CASSANDRA_HOSTS")
	var hosts []string
	if hostsStr != "" {
		hosts = strings.Split(hostsStr, ",")
		// Opcional: remover espacios en blanco en cada host
		for i, host := range hosts {
			hosts[i] = strings.TrimSpace(host)
		}
	}

	// Obtener el resto de variables de entorno
	keyspace := os.Getenv("CASSANDRA_KEYSPACE")
	username := os.Getenv("CASSANDRA_USERNAME")
	password := os.Getenv("CASSANDRA_PASSWORD")

	config := newConfig(hosts, keyspace, username, password)

	if err := config.Validate(); err != nil {
		return nil, err
	}

	return newRepository(config)
}
