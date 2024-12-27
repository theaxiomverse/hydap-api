// pkg/modules/core/metrics.go
package core

import (
	"github.com/prometheus/client_golang/prometheus"
	"sync"
)

type MetricsExporter struct {
	registry *prometheus.Registry
	modules  map[string]*moduleMetrics
	mu       sync.RWMutex
}

type moduleMetrics struct {
	health   prometheus.Gauge
	memory   prometheus.Gauge
	uptime   prometheus.Counter
	requests prometheus.Counter
}

func NewMetricsExporter() *MetricsExporter {
	return &MetricsExporter{
		registry: prometheus.NewRegistry(),
		modules:  make(map[string]*moduleMetrics),
	}
}

func (me *MetricsExporter) RegisterModule(name string) {
	me.mu.Lock()
	defer me.mu.Unlock()

	mm := &moduleMetrics{
		health: prometheus.NewGauge(prometheus.GaugeOpts{
			Name:        "module_health",
			Help:        "Module health status",
			ConstLabels: prometheus.Labels{"module": name},
		}),
		memory: prometheus.NewGauge(prometheus.GaugeOpts{
			Name:        "module_memory_bytes",
			Help:        "Module memory usage in bytes",
			ConstLabels: prometheus.Labels{"module": name},
		}),
		uptime: prometheus.NewCounter(prometheus.CounterOpts{
			Name:        "module_uptime_seconds",
			Help:        "Module uptime in seconds",
			ConstLabels: prometheus.Labels{"module": name},
		}),
		requests: prometheus.NewCounter(prometheus.CounterOpts{
			Name:        "module_requests_total",
			Help:        "Total number of module requests",
			ConstLabels: prometheus.Labels{"module": name},
		}),
	}

	me.registry.MustRegister(mm.health, mm.memory, mm.uptime, mm.requests)
	me.modules[name] = mm
}

func (me *MetricsExporter) Modules() map[string]*moduleMetrics {
	me.mu.RLock()
	defer me.mu.RUnlock()
	return me.modules
}
