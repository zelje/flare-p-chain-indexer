package shared

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

type metrics struct {
	// Last accepted index on the chain
	lastAcceptedIndex prometheus.Gauge

	// Last processed index by the indexer
	lastProcessedIndex prometheus.Gauge

	// Processing time in milliseconds
	processingTime prometheus.Gauge
}

func newMetrics(namespace string) *metrics {
	return &metrics{
		lastAcceptedIndex: promauto.NewGauge(prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "last_accepted_index",
			Help:      "Last accepted index on the chain",
		}),
		lastProcessedIndex: promauto.NewGauge(prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "last_processed_index",
			Help:      "Last processed index by the indexer",
		}),
		processingTime: promauto.NewGauge(prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "last_processing_time",
			Help:      "Time of processing of the last batch in milliseconds",
		}),
	}
}

func (m *metrics) Update(newAcceptedCount uint64, newProcessedCount uint64, processingTime int64) {
	m.lastAcceptedIndex.Set(float64(newAcceptedCount))
	m.lastProcessedIndex.Set(float64(newProcessedCount))
	m.processingTime.Set(float64(processingTime))
}
