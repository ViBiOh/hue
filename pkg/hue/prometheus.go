package hue

import (
	"fmt"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
)

func (a *app) getMetrics(prefix, suffix string) prometheus.Gauge {
	if gauge, ok := a.prometheusCollectors[fmt.Sprintf("%s_%s", prefix, suffix)]; ok {
		return gauge
	}

	gauge := prometheus.NewGauge(prometheus.GaugeOpts{
		Name: fmt.Sprintf("hue_%s_%s", prefix, suffix),
	})

	a.prometheusCollectors[fmt.Sprintf("%s_%s", prefix, suffix)] = gauge
	a.prometheusRegisterer.MustRegister(gauge)

	return gauge
}

func (a *app) updatePrometheusSensors() {
	if a.prometheusRegisterer == nil {
		return
	}

	a.mutex.RLock()
	defer a.mutex.RUnlock()

	for _, sensor := range a.sensors {
		a.getMetrics(strings.ToLower(sensor.Name), "temperature").Set(float64(sensor.State.Temperature))
		a.getMetrics(strings.ToLower(sensor.Name), "battery").Set(float64(sensor.Config.Battery))
	}
}
