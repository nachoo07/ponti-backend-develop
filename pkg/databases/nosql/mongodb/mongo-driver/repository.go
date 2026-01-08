package pkgmongo

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	instance  Repository
	once      sync.Once
	initError error
)

type repository struct {
	db *mongo.Database
}

func newRepository(c Config) (Repository, error) {
	once.Do(func() {
		instance = &repository{}
		initError = instance.Connect(c)
		if initError != nil {
			instance = nil
		}
	})
	return instance, initError
}

func (r *repository) Connect(c Config) error {
	dsn := c.DSN()
	clientOptions := options.Client().ApplyURI(dsn)

	conn, err := mongo.Connect(context.TODO(), clientOptions)
	if err != nil {
		return fmt.Errorf("failed to connect to MongoDB: %w", err)
	}

	// Verificar la conexión con un ping
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err = conn.Ping(ctx, nil); err != nil {
		return fmt.Errorf("failed to ping MongoDB: %w", err)
	}

	r.db = conn.Database(c.Database())
	log.Printf("MongoDB successfully connected to database: %s", c.Database())
	return nil
}

func (r *repository) Close() {
	if r.db != nil {
		r.db.Client().Disconnect(context.TODO())
	}
}

func (r *repository) DB() *mongo.Database {
	return r.db
}
