module paymentservice

go 1.25.3

replace gitea.xscloud.ru/xscloud/golib v1.2.2 => github.com/veresnikov/rp-golib v1.2.2

require (
	gitea.xscloud.ru/xscloud/golib v1.2.2
	github.com/google/uuid v1.6.0
	github.com/gorilla/mux v1.7.4
	github.com/kelseyhightower/envconfig v1.4.0
	github.com/pkg/errors v0.9.1
	github.com/urfave/cli/v2 v2.27.7
	golang.org/x/sync v0.12.0
	google.golang.org/grpc v1.69.4
	google.golang.org/protobuf v1.35.1
)

require (
	filippo.io/edwards25519 v1.1.0 // indirect
	github.com/cenkalti/backoff v2.2.1+incompatible // indirect
	github.com/cpuguy83/go-md2man/v2 v2.0.7 // indirect
	github.com/go-sql-driver/mysql v1.9.3 // indirect
	github.com/google/go-cmp v0.7.0 // indirect
	github.com/jmoiron/sqlx v1.4.0 // indirect
	github.com/rabbitmq/amqp091-go v1.10.0 // indirect
	github.com/russross/blackfriday/v2 v2.1.0 // indirect
	github.com/sirupsen/logrus v1.9.3 // indirect
	github.com/stretchr/testify v1.10.0 // indirect
	github.com/xrash/smetrics v0.0.0-20240521201337-686a1a2994c1 // indirect
	go.opentelemetry.io/otel v1.37.0 // indirect
	golang.org/x/net v0.38.0 // indirect
	golang.org/x/sys v0.31.0 // indirect
	golang.org/x/text v0.23.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20241015192408-796eee8c2d53 // indirect
)
