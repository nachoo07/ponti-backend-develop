package pkgcassandra

import (
	"fmt"
)



type config struct {
	hosts    []string
	keyspace string
	username string
	password string
}

func newConfig(hosts []string, keyspace string, username string, password string) Config {
	h := make([]string, len(hosts))
	copy(h, hosts)
	return &config{
		hosts:    h,
		keyspace: keyspace,
		username: username,
		password: password,
	}
}

func (c *config) GetHosts() []string {
	return c.hosts
}

func (c *config) SetHosts(hosts []string) {
	h := make([]string, len(hosts))
	copy(h, hosts)
	c.hosts = h
}

func (c *config) GetKeyspace() string {
	return c.keyspace
}

func (c *config) SetKeyspace(keyspace string) {
	c.keyspace = keyspace
}

func (c *config) GetUsername() string {
	return c.username
}

func (c *config) SetUsername(username string) {
	c.username = username
}

func (c *config) GetPassword() string {
	return c.password
}

func (c *config) SetPassword(password string) {
	c.password = password
}

func (c *config) Validate() error {
	if len(c.hosts) == 0 {
		return fmt.Errorf("cassandra hosts are not configured")
	}
	if c.keyspace == "" {
		return fmt.Errorf("cassandra keyspace is not configured")
	}
	if c.username == "" {
		return fmt.Errorf("cassandra username is not configured")
	}
	if c.password == "" {
		return fmt.Errorf("cassandra password is not configured")
	}
	return nil
}
