package v2

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

func (a *App) createMetrics(meterProvider metric.MeterProvider) error {
	if meterProvider == nil {
		return nil
	}

	meter := meterProvider.Meter("github.com/ViBiOh/hue/pkg/v2")

	if err := a.createTemperatureMetric(meter); err != nil {
		return fmt.Errorf("create temperature metric: %w", err)
	}

	if err := a.createBatteryMetric(meter); err != nil {
		return fmt.Errorf("create battery metric: %w", err)
	}

	if err := a.createMotionMetric(meter); err != nil {
		return fmt.Errorf("create motion metric: %w", err)
	}

	return nil
}

func (a *App) createTemperatureMetric(meter metric.Meter) error {
	_, err := meter.Float64ObservableGauge("hue.temperature", metric.WithFloat64Callback(func(ctx context.Context, fo metric.Float64Observer) error {
		a.mutex.RLock()
		defer a.mutex.RUnlock()

		for _, motion := range a.motionSensors {
			fo.Observe(motion.Temperature, metric.WithAttributes(
				attribute.String("room", motion.Name),
			))
		}

		return nil
	}))

	return err
}

func (a *App) createBatteryMetric(meter metric.Meter) error {
	_, err := meter.Int64ObservableGauge("hue.battery", metric.WithInt64Callback(func(ctx context.Context, io metric.Int64Observer) error {
		a.mutex.RLock()
		defer a.mutex.RUnlock()

		for _, motion := range a.motionSensors {
			io.Observe(motion.BatteryLevel, metric.WithAttributes(
				attribute.String("room", motion.Name),
			))
		}

		return nil
	}))

	return err
}

func (a *App) createMotionMetric(meter metric.Meter) error {
	_, err := meter.Int64ObservableGauge("hue.motion", metric.WithInt64Callback(func(ctx context.Context, io metric.Int64Observer) error {
		a.mutex.RLock()
		defer a.mutex.RUnlock()

		for _, motion := range a.motionSensors {
			var value int64
			if motion.Motion {
				value = 1
			}

			io.Observe(value, metric.WithAttributes(
				attribute.String("room", motion.Name),
			))
		}

		return nil
	}))

	return err
}
