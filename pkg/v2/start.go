package v2

import (
	"context"
	"fmt"
	"log/slog"
	"os"
)

func (s *Service) Start(ctx context.Context) {
	if err := s.initConfig(ctx); err != nil {
		slog.ErrorContext(ctx, "init", "err", err)
		os.Exit(1)
	}

	go s.streamIndefinitely(ctx.Done())
}

func (s *Service) initConfig(ctx context.Context) (err error) {
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
