package dyson

import (
	"fmt"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
)

func (a *App) getMetrics(prefix, suffix string) prometheus.Gauge {
	gauge, ok := a.prometheusCollectors[fmt.Sprintf(`%s_%s`, prefix, suffix)]
	if !ok {
		gauge = prometheus.NewGauge(prometheus.GaugeOpts{
			Name: fmt.Sprintf(`%s_%s_%s`, strings.ToLower(Source), prefix, suffix),
		})

		a.prometheusCollectors[fmt.Sprintf(`%s_%s`, prefix, suffix)] = gauge
		prometheus.MustRegister(gauge)
	}

	return gauge
}

func (a *App) updatePrometheus() {
	for _, device := range a.devices {
		a.getMetrics(strings.ToLower(device.Name), `temperature`).Set(float64(device.State.Temperature))
		a.getMetrics(strings.ToLower(device.Name), `humidity`).Set(float64(device.State.Humidity))
	}
}
