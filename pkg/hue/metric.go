package hue

import (
	"fmt"

	"github.com/prometheus/client_golang/prometheus"
)

func createMetrics(prometheusRegisterer prometheus.Registerer, names ...string) (map[string]*prometheus.GaugeVec, error) {
	if prometheusRegisterer == nil {
		return nil, nil
	}

	metrics := make(map[string]*prometheus.GaugeVec)
	for _, name := range names {
		metric, err := createMetric(prometheusRegisterer, name)
		if err != nil {
			return nil, err
		}

		metrics[name] = metric
	}

	return metrics, nil
}

func createMetric(prometheusRegisterer prometheus.Registerer, name string) (*prometheus.GaugeVec, error) {
	if prometheusRegisterer == nil {
		return nil, nil
	}

	gauge := prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "hue",
		Name:      name,
	}, []string{"room"})

	if err := prometheusRegisterer.Register(gauge); err != nil {
		return nil, fmt.Errorf("unable to registrer %s: %s", name, err)
	}

	return gauge, nil
}

func (a *App) setMetric(name, room string, value float64) {
	metric, ok := a.metrics[name]
	if !ok {
		return
	}

	labels := prometheus.Labels{
		"room": room,
	}

	metric.With(labels).Set(value)
}

func (a *App) updatePrometheus() {
	a.mutex.RLock()
	defer a.mutex.RUnlock()

	for _, sensor := range a.sensors {
		a.setMetric("temperature", sensor.Name, float64(sensor.State.Temperature))
		a.setMetric("battery", sensor.Name, float64(sensor.Config.Battery))
	}
}
