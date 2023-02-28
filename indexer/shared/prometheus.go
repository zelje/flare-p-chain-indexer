package shared

import (
	"flare-indexer/indexer/config"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func InitMetricsServer(cfg *config.MetricsConfig) {
	if len(cfg.PrometheusAddress) == 0 {
		return
	}

	r := mux.NewRouter()

	r.Path("/metrics").Handler(promhttp.Handler())

	srv := &http.Server{
		Addr:    cfg.PrometheusAddress,
		Handler: r,
	}
	go srv.ListenAndServe()
}
