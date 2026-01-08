package pkgcassandra

import (
	"fmt"
	"log"
	"sync"

	"github.com/gocql/gocql"
)

var (
	instance  Repository
	once      sync.Once
	initError error
)

type repository struct {
	session *gocql.Session
}

func newRepository(config Config) (Repository, error) {
	once.Do(func() {
		// Asignamos directamente a la variable global "instance", sin redeclararla.
		instance = &repository{}
		initError = instance.Connect(config)
		if initError != nil {
			instance = nil
		}
	})
	return instance, initError
}

func (c *repository) Connect(config Config) error {
	// Conectar sin especificar keyspace para poder crear el keyspace si no existe.
	cluster := gocql.NewCluster(config.GetHosts()...)
	cluster.Authenticator = gocql.PasswordAuthenticator{
		Username: config.GetUsername(),
		Password: config.GetPassword(),
	}
	// Conectar sin keyspace
	session, err := cluster.CreateSession()
	if err != nil {
		return fmt.Errorf("failed to connect to Cassandra: %w", err)
	}
	// Cerrar la sesión temporal al finalizar.
	defer session.Close()

	// Crear el keyspace si no existe.
	// Se usa el nombre proporcionado en config.GetKeyspace()
	createKeyspaceCQL := fmt.Sprintf(
		`CREATE KEYSPACE IF NOT EXISTS %s WITH replication = {'class': 'SimpleStrategy', 'replication_factor': 3}`,
		config.GetKeyspace(),
	)
	if err := session.Query(createKeyspaceCQL).Exec(); err != nil {
		return fmt.Errorf("failed to create keyspace %s: %w", config.GetKeyspace(), err)
	}

	// Ahora reconectar especificando el keyspace creado.
	cluster.Keyspace = config.GetKeyspace()
	sessionWithKeyspace, err := cluster.CreateSession()
	if err != nil {
		return fmt.Errorf("failed to connect to Cassandra keyspace: %w", err)
	}
	c.session = sessionWithKeyspace

	log.Printf("Cassandra successfully connected to keyspace: %s", config.GetKeyspace())

	return nil
}

func (c *repository) Close() {
	if c.session != nil {
		c.session.Close()
	}
}

func (c *repository) GetSession() *gocql.Session {
	return c.session
}
