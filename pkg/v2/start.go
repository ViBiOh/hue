package v2

import (
	"context"
	"fmt"
	"log/slog"
)

func (s *Service) Start(ctx context.Context) {
	s.streamIndefinitely(ctx.Done())
}

func (s *Service) Init(ctx context.Context) (err error) {
	slog.Info("Initializing V2...")
	defer slog.Info("Initialization V2 done.")

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
