package main

import (
	"errors"
	"net/http"
	"time"

	"gitea.xscloud.ru/xscloud/golib/pkg/application/logging"
	libio "gitea.xscloud.ru/xscloud/golib/pkg/common/io"
	"gitea.xscloud.ru/xscloud/golib/pkg/infrastructure/amqp"
	"gitea.xscloud.ru/xscloud/golib/pkg/infrastructure/mysql"
	"github.com/gorilla/mux"
	"github.com/urfave/cli/v2"
	"golang.org/x/sync/errgroup"

	"notificationservice/pkg/notification/infrastructure/consumer"
)

type messageHandlerConfig struct {
	Service  Service  `envconfig:"service"`
	Database Database `envconfig:"database" required:"true"`
	AMQP     AMQP     `envconfig:"amqp" required:"true"`
}

func messageHandler(logger logging.Logger) *cli.Command {
	return &cli.Command{
		Name:   "message-handler",
		Before: migrateImpl(logger),
		Action: func(c *cli.Context) error {
			cnf, err := parseEnvs[messageHandlerConfig]()
			if err != nil {
				return err
			}

			closer := libio.NewMultiCloser()
			defer func() {
				err = errors.Join(err, closer.Close())
			}()

			databaseConnector, err := newDatabaseConnector(cnf.Database)
			if err != nil {
				return err
			}
			closer.AddCloser(databaseConnector)
			databaseConnectionPool := mysql.NewConnectionPool(databaseConnector.TransactionalClient())

			amqpConnection := newAMQPConnection(cnf.AMQP, logger)

			eventConsumer, err := consumer.NewEventConsumer(c.Context, amqpConnection, databaseConnectionPool, logger)
			if err != nil {
				return err
			}

			queueConfig := &amqp.QueueConfig{
				Name:    "notification_events",
				Durable: true,
			}
			bindConfig := &amqp.BindConfig{
				QueueName:    "notification_events",
				ExchangeName: "domain_event_exchange",
				RoutingKeys:  []string{"order.*", "user.*"},
			}

			amqpConnection.Consumer(
				c.Context,
				eventConsumer.Handler(),
				queueConfig,
				bindConfig,
				nil,
			)

			err = amqpConnection.Start()
			if err != nil {
				return err
			}
			closer.AddCloser(libio.CloserFunc(func() error {
				return amqpConnection.Stop()
			}))

			errGroup := errgroup.Group{}

			errGroup.Go(func() error {
				router := mux.NewRouter()
				registerHealthcheck(router)
				registerMetrics(router)
				server := http.Server{
					Addr:              cnf.Service.HTTPAddress,
					Handler:           router,
					ReadHeaderTimeout: 5 * time.Second,
				}
				graceCallback(c.Context, logger, cnf.Service.GracePeriod, server.Shutdown)
				return server.ListenAndServe()
			})

			return errGroup.Wait()
		},
	}
}
