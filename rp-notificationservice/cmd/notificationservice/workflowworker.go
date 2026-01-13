package main

import (
	"net/http"
	"time"

	"gitea.xscloud.ru/xscloud/golib/pkg/application/logging"
	"github.com/gorilla/mux"
	"github.com/urfave/cli/v2"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
	"golang.org/x/sync/errgroup"

	"notificationservice/pkg/notification/infrastructure/temporal/activity"
)

type workflowWorkerConfig struct {
	Service  Service  `envconfig:"service"`
	Temporal Temporal `envconfig:"temporal" required:"true"`
}

func workflowWorker(logger logging.Logger) *cli.Command {
	return &cli.Command{
		Name: "workflow-worker",
		Action: func(c *cli.Context) error {
			cnf, err := parseEnvs[workflowWorkerConfig]()
			if err != nil {
				return err
			}

			temporalClient, err := client.Dial(client.Options{
				HostPort: cnf.Temporal.Host,
			})
			if err != nil {
				return err
			}
			defer temporalClient.Close()

			w := worker.New(temporalClient, "notificationservice_task_queue", worker.Options{})

			activities := &activity.NotificationActivities{}
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
