package main

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func registerMetrics(router *mux.Router) {
	router.Handle("/metrics", promhttp.Handler()).Methods(http.MethodGet)
}
