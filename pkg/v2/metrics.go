package v2

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

const motionValue = 1

func (s *Service) createMetrics(meterProvider metric.MeterProvider) error {
	if meterProvider == nil {
		return nil
	}

	meter := meterProvider.Meter("github.com/ViBiOh/hue/pkg/v2")

	if err := s.createTemperatureMetric(meter); err != nil {
		return fmt.Errorf("create temperature metric: %w", err)
	}

	if err := s.createBatteryMetric(meter); err != nil {
		return fmt.Errorf("create battery metric: %w", err)
	}

	if err := s.createMotionMetric(meter); err != nil {
		return fmt.Errorf("create motion metric: %w", err)
	}

	return nil
}

func (s *Service) createTemperatureMetric(meter metric.Meter) error {
	_, err := meter.Float64ObservableGauge("hue.temperature", metric.WithFloat64Callback(func(ctx context.Context, fo metric.Float64Observer) error {
		s.mutex.RLock()
		defer s.mutex.RUnlock()

		for _, motion := range s.motionSensors {
			fo.Observe(motion.Temperature, metric.WithAttributes(
				attribute.String("room", motion.Name),
			))
		}

		return nil
	}))

	return err
}

func (s *Service) createBatteryMetric(meter metric.Meter) error {
	_, err := meter.Int64ObservableGauge("hue.battery", metric.WithInt64Callback(func(ctx context.Context, io metric.Int64Observer) error {
		s.mutex.RLock()
		defer s.mutex.RUnlock()

		for _, motion := range s.motionSensors {
			io.Observe(motion.BatteryLevel, metric.WithAttributes(
				attribute.String("room", motion.Name),
			))
		}

		return nil
	}))

	return err
}

func (s *Service) createMotionMetric(meter metric.Meter) error {
	_, err := meter.Int64ObservableGauge("hue.motion", metric.WithInt64Callback(func(ctx context.Context, io metric.Int64Observer) error {
		s.mutex.RLock()
		defer s.mutex.RUnlock()

		for _, motion := range s.motionSensors {
			var value int64
			if motion.Motion {
				value = motionValue
			}

			io.Observe(value, metric.WithAttributes(
				attribute.String("room", motion.Name),
			))
		}

		return nil
	}))

	return err
}
