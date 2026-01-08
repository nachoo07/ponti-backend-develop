package pkgmongo

import "os"

func Bootstrap(user, password, host, port, databaseName string) (Repository, error) {

	if user == "" {
		user = os.Getenv("MONGO_USER")
	}
	if password == "" {
		password = os.Getenv("MONGO_PASSWORD")
	}
	if host == "" {
		host = os.Getenv("MONGO_HOST")
	}
	if port == "" {
		port = os.Getenv("MONGO_PORT")
	}
	if databaseName == "" {
		databaseName = os.Getenv("MONGO_DATABASE")
	}

	config := newConfig(
		user,
		password,
		host,
		port,
		databaseName,
	)

	if err := config.Validate(); err != nil {
		return nil, err
	}

	return newRepository(config)
}
