package main

import (
	"net/http"

	"gitea.xscloud.ru/xscloud/golib/pkg/application/logging"
	libio "gitea.xscloud.ru/xscloud/golib/pkg/common/io"
	"gitea.xscloud.ru/xscloud/golib/pkg/infrastructure/amqp"
	"gitea.xscloud.ru/xscloud/golib/pkg/infrastructure/mysql"
	"gitea.xscloud.ru/xscloud/golib/pkg/infrastructure/outbox"
	"github.com/gorilla/mux"
	"github.com/urfave/cli/v2"
	"golang.org/x/sync/errgroup"

	appservice "userservice/pkg/user/application/service"
	"userservice/pkg/user/infrastructure/integrationevent"
	inframysql "userservice/pkg/user/infrastructure/mysql"
	"userservice/pkg/user/infrastructure/temporal"
)

type messageHandlerConfig struct {
	Service  Service  `envconfig:"service"`
	Database Database `envconfig:"database" required:"true"`
	AMQP     AMQP     `envconfig:"amqp" required:"true"`
	Temporal Temporal `envconfig:"temporal" required:"true"`
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
			defer func() { _ = closer.Close() }()

			databaseConnector, err := newDatabaseConnector(cnf.Database)
			if err != nil {
				return err
			}
			closer.AddCloser(databaseConnector)
			databaseConnectionPool := mysql.NewConnectionPool(databaseConnector.TransactionalClient())

			temporalClient, err := temporal.NewClient(logger, cnf.Temporal.Host)
			if err != nil {
				return err
			}
			closer.AddCloser(libio.CloserFunc(func() error {
				temporalClient.Close()
				return nil
			}))
			workflowService := temporal.NewWorkflowService(temporalClient)

			amqpConnection := newAMQPConnection(cnf.AMQP, logger)

			amqpEventProducer := amqpConnection.Producer(
				&amqp.ExchangeConfig{
					Name:    integrationevent.ExchangeName,
					Kind:    integrationevent.ExchangeKind,
					Durable: true,
				},
				nil,
				nil,
			)

			libUoW := mysql.NewUnitOfWork(databaseConnectionPool, inframysql.NewRepositoryProvider)
			libLUow := mysql.NewLockableUnitOfWork(libUoW, mysql.NewLocker(databaseConnectionPool))
			uow := inframysql.NewUnitOfWork(libUoW)
			luow := inframysql.NewLockableUnitOfWork(libLUow)

			eventDispatcher := outbox.NewEventDispatcher(
				appID,
				integrationevent.TransportName,
				integrationevent.NewEventSerializer(),
				libUoW,
			)
			userService := appservice.NewUserService(uow, luow, eventDispatcher)

			amqpTransport := integrationevent.NewAMQPTransport(logger, workflowService, userService)

			amqpConnection.Consumer(
				c.Context,
				amqpTransport.Handler(),
				&amqp.QueueConfig{Name: integrationevent.QueueName, Durable: true},
				&amqp.BindConfig{
					QueueName:    integrationevent.QueueName,
					ExchangeName: integrationevent.ExchangeName,
					RoutingKeys:  []string{integrationevent.RoutingKeyPrefix + "#"},
				},
				nil,
			)

			if err = amqpConnection.Start(); err != nil {
				return err
			}
			closer.AddCloser(libio.CloserFunc(func() error {
				return amqpConnection.Stop()
			}))

			outboxEventHandler := outbox.NewEventHandler(outbox.EventHandlerConfig{
				TransportName:  integrationevent.TransportName,
				Transport:      integrationevent.NewTransport(logger, amqpEventProducer),
				ConnectionPool: databaseConnectionPool,
				Logger:         logger,
			})

			errGroup := errgroup.Group{}
			errGroup.Go(func() error {
				return outboxEventHandler.Start(c.Context)
			})

			errGroup.Go(func() error {
				router := mux.NewRouter()
				registerHealthcheck(router)
				server := http.Server{
					Addr:    cnf.Service.HTTPAddress,
					Handler: router,
				}
				graceCallback(c.Context, logger, cnf.Service.GracePeriod, server.Shutdown)
				return server.ListenAndServe()
			})

			return errGroup.Wait()
		},
	}
}
