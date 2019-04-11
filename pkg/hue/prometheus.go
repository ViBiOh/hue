package hue

import (
	"fmt"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
)

func (a *App) getMetrics(prefix, suffix string) prometheus.Gauge {
	gauge, ok := a.prometheusCollectors[fmt.Sprintf("%s_%s", prefix, suffix)]
	if !ok {
		gauge = prometheus.NewGauge(prometheus.GaugeOpts{
			Name: fmt.Sprintf("%s_%s_%s", strings.ToLower(Source), prefix, suffix),
		})

		a.prometheusCollectors[fmt.Sprintf("%s_%s", prefix, suffix)] = gauge
		prometheus.MustRegister(gauge)
	}

	return gauge
}

func (a *App) updatePrometheusSensors() {
	for _, sensor := range a.sensors {
		a.getMetrics(strings.ToLower(sensor.Name), "temperature").Set(float64(sensor.State.Temperature))
		a.getMetrics(strings.ToLower(sensor.Name), "battery").Set(float64(sensor.Config.Battery))
	}
}
