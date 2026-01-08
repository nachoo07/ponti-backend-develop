package pkgmongo

import "go.mongodb.org/mongo-driver/mongo"

type Repository interface {
	Connect(Config) error
	Close()
	DB() *mongo.Database
}

type Config interface {
	Validate() error
	DSN() string
	Database() string
}
