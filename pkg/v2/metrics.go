package v2

import (
	"fmt"

	"go.opentelemetry.io/otel/metric"
)

const motionValue = 1

func (s *Service) createMetrics(meterProvider metric.MeterProvider) error {
	if meterProvider == nil {
		return nil
	}

	meter := meterProvider.Meter("github.com/ViBiOh/hue/pkg/v2")

	var err error

	s.temperatureMetric, err = meter.Float64Gauge("hue.temperature")
	if err != nil {
		return fmt.Errorf("create temperature metric: %w", err)
	}

	s.batteryMetric, err = meter.Int64Gauge("hue.battery")
	if err != nil {
		return fmt.Errorf("create battery metric: %w", err)
	}

	s.motionMetric, err = meter.Int64Gauge("hue.motion")
	if err != nil {
		return fmt.Errorf("create motion metric: %w", err)
	}

	s.lightLevelMetric, err = meter.Int64Gauge("hue.light_level")
	if err != nil {
		return fmt.Errorf("create light level metric: %w", err)
	}

	return nil
}
