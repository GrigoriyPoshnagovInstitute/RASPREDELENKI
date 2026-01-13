package main

import (
	"net/http"
	"time"

	"gitea.xscloud.ru/xscloud/golib/pkg/application/logging"
	libio "gitea.xscloud.ru/xscloud/golib/pkg/common/io"
	"github.com/gorilla/mux"
	"github.com/urfave/cli/v2"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
	"golang.org/x/sync/errgroup"

	"productservice/pkg/product/infrastructure/mysql/query"
	"productservice/pkg/product/infrastructure/temporal/activity"
)

type workflowWorkerConfig struct {
	Service  Service  `envconfig:"service"`
	Database Database `envconfig:"database" required:"true"`
	Temporal Temporal `envconfig:"temporal" required:"true"`
}

func workflowWorker(logger logging.Logger) *cli.Command {
	return &cli.Command{
		Name:   "workflow-worker",
		Before: migrateImpl(logger),
		Action: func(c *cli.Context) error {
			cnf, err := parseEnvs[workflowWorkerConfig]()
			if err != nil {
				return err
			}

			closer := libio.NewMultiCloser()
			defer func() {
				_ = closer.Close()
			}()

			databaseConnector, err := newDatabaseConnector(cnf.Database)
			if err != nil {
				return err
			}
			closer.AddCloser(databaseConnector)

			temporalClient, err := client.Dial(client.Options{
				HostPort: cnf.Temporal.Host,
			})
			if err != nil {
				return err
			}
			closer.AddCloser(libio.CloserFunc(func() error {
				temporalClient.Close()
				return nil
			}))

			productQueryService := query.NewProductQueryService(databaseConnector.TransactionalClient())

			w := worker.New(temporalClient, "productservice_task_queue", worker.Options{})

			activities := activity.NewProductActivities(productQueryService)
			w.RegisterActivity(activities)

			errGroup := errgroup.Group{}
			errGroup.Go(func() error {
				return w.Run(worker.InterruptCh())
			})

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
