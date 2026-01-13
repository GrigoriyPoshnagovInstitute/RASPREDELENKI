package main

import (
	"time"

	"github.com/kelseyhightower/envconfig"
	"github.com/pkg/errors"
)

func parseEnvs[T any]() (T, error) {
	var c T
	err := envconfig.Process(appID, &c)
	return c, errors.WithStack(err)
}

type Service struct {
	GracePeriod time.Duration `envconfig:"GRACE_PERIOD" default:"15s"`
	GRPCAddress string        `envconfig:"GRPC_ADDRESS" default:":8081"`
	HTTPAddress string        `envconfig:"HTTP_ADDRESS" default:":8082"`
}

type Database struct {
	User                  string        `envconfig:"USER" required:"true"`
	Password              string        `envconfig:"PASSWORD" required:"true"`
	Host                  string        `envconfig:"HOST" required:"true"`
	Name                  string        `envconfig:"NAME" required:"true"`
	MaxConnections        int           `envconfig:"MAX_CONNECTIONS" default:"20"`
	ConnectionMaxLifeTime time.Duration `envconfig:"CONNECTION_MAX_LIFE_TIME" default:"10m"`
	ConnectionMaxIdleTime time.Duration `envconfig:"CONNECTION_MAX_IDLE_TIME" default:"1m"`
}

type AMQP struct {
	User           string        `envconfig:"USER" required:"true"`
	Password       string        `envconfig:"PASSWORD" required:"true"`
	Host           string        `envconfig:"HOST" required:"true"`
	ConnectTimeout time.Duration `envconfig:"CONNECT_TIMEOUT"`
}

type Temporal struct {
	Host string `envconfig:"HOST" required:"true"`
}
