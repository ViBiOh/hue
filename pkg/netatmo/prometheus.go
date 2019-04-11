package netatmo

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
		a.getMetrics(strings.ToLower(device.StationName), "temperature").Set(float64(device.DashboardData.Temperature))
		a.getMetrics(strings.ToLower(device.StationName), "humidity").Set(float64(device.DashboardData.Humidity))
		a.getMetrics(strings.ToLower(device.StationName), "noise").Set(float64(device.DashboardData.Noise))
		a.getMetrics(strings.ToLower(device.StationName), "co2").Set(float64(device.DashboardData.CO2))

		for _, module := range device.Modules {
			a.getMetrics(strings.ToLower(fmt.Sprintf("%s_%s", device.StationName, module.ModuleName)), "temperature").Set(float64(module.DashboardData.Temperature))
			a.getMetrics(strings.ToLower(fmt.Sprintf("%s_%s", device.StationName, module.ModuleName)), "humidity").Set(float64(module.DashboardData.Humidity))

		}
	}
}
