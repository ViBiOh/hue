package v2

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"runtime"
)

func (s *Service) Start(ctx context.Context) {
	s.streamIndefinitely(ctx.Done())
}

func (s *Service) Init(ctx context.Context) (err error) {
	slog.Info("Initializing V2...")
	defer slog.Info("Initialization V2 done.")

	devices := make(chan Device, runtime.NumCPU())
	var streamErr error

	go func() {
		if err := s.streamDevices(ctx, devices); err != nil {
			streamErr = fmt.Errorf("build lights: %w", err)
		}
	}()

	defer func() {
		err = errors.Join(err, streamErr)
	}()

	s.taps, err = s.buildTaps(ctx, devices)
	if err != nil {
		err = fmt.Errorf("build tap: %w", err)
		return
	}

	s.lights, err = s.buildLights(ctx)
	if err != nil {
		err = fmt.Errorf("build lights: %w", err)
		return
	}

	s.groups, err = s.buildGroup(ctx)
	if err != nil {
		err = fmt.Errorf("build groups: %w", err)
		return
	}

	s.motionSensors, err = s.buildMotionSensor(ctx)
	if err != nil {
		err = fmt.Errorf("build motion sensor: %w", err)
		return
	}

	return nil
}
