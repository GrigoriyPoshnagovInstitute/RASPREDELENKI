package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	DatabaseDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: "product",
		Subsystem: "database",
		Name:      "query_duration_seconds",
		Help:      "Duration of database queries",
	}, []string{"operation", "table", "status"})
)
