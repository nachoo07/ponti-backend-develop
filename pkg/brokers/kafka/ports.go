package pkgafka

import "context"

type Config interface {
	GetBrokers() []string
	GetGroupID() string
	Validate() error
}

type Service interface {
	Publish(context.Context, string, []byte, []byte) error
	Consume(context.Context, []string, func([]byte, []byte) error) error
}
