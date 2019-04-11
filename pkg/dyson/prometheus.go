package dyson

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

func (a *App) updatePrometheus() {
	for _, device := range a.devices {
		deviceName := strings.ToLower(device.Name)
		deviceName = strings.Replace(deviceName, " ", "_", -1)
		deviceName = strings.Replace(deviceName, "+", "_", -1)

		a.getMetrics(deviceName, "temperature").Set(float64(device.State.Temperature))
		a.getMetrics(deviceName, "humidity").Set(float64(device.State.Humidity))
	}
}
