package telemetry

import (
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

type PlanMetrics struct {
	planDuration *prometheus.HistogramVec
}

type PlanMetricsRecorder interface {
	ObservePlanGeneration(d time.Duration, cluster, namespace string)
}

var (
	planDurationName    = "plan_generation_duration_seconds"
	planDurationHelp    = "End-to-end duration for generating plan previews from prompts"
	planDurationBuckets = []float64{0.05, 0.1, 0.25, 0.5, 1, 2, 3, 5}
)

type collectorPair struct {
	hist *prometheus.HistogramVec
}

var (
	collectorCache = struct {
		sync.Mutex
		m map[prometheus.Registerer]collectorPair
	}{m: make(map[prometheus.Registerer]collectorPair)}
)

func NewPlanMetrics(registerer prometheus.Registerer) *PlanMetrics {
	if registerer == nil {
		registerer = prometheus.DefaultRegisterer
	}

	collectorCache.Lock()
	defer collectorCache.Unlock()

	if pair, ok := collectorCache.m[registerer]; ok {
		return &PlanMetrics{planDuration: pair.hist}
	}

	histogram := prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    planDurationName,
			Help:    planDurationHelp,
			Buckets: planDurationBuckets,
		},
		[]string{"cluster", "namespace"},
	)

	if err := registerer.Register(histogram); err != nil {
		if already, ok := err.(prometheus.AlreadyRegisteredError); ok {
			histogram = already.ExistingCollector.(*prometheus.HistogramVec)
		} else {
			panic(err)
		}
	}

	collectorCache.m[registerer] = collectorPair{hist: histogram}
	return &PlanMetrics{planDuration: histogram}
}

func (m *PlanMetrics) ObservePlanGeneration(d time.Duration, cluster, namespace string) {
	if m == nil {
		return
	}
	if cluster == "" {
		cluster = "unknown"
	}
	if namespace == "" {
		namespace = "unspecified"
	}
	m.planDuration.WithLabelValues(cluster, namespace).Observe(d.Seconds())
}
