package enedis

import (
	"fmt"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
)

func (a *App) getMetrics(name string) prometheus.Gauge {
	gauge, ok := a.prometheusCollectors[name]
	if !ok {
		gauge = prometheus.NewGauge(prometheus.GaugeOpts{
			Name: fmt.Sprintf(`%s_%s`, strings.ToLower(Source), name),
		})

		a.prometheusCollectors[name] = gauge
		prometheus.MustRegister(gauge)
	}

	return gauge
}

func (a *App) updatePrometheusSensors() {
	a.getMetrics(`temperature`).Set(2.543)
}
