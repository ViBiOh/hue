package v2

import (
	"context"
	"fmt"

	"github.com/ViBiOh/httputils/v4/pkg/logger"
)

// Start worker
func (a *App) Start(ctx context.Context) {
	if err := a.initConfig(ctx); err != nil {
		logger.Fatal(err)
	}

	go a.streamIndefinitely(ctx.Done())
}

func (a *App) initConfig(ctx context.Context) (err error) {
	a.lights, err = a.buildLights(ctx)
	if err != nil {
		err = fmt.Errorf("build lights: %w", err)
		return
	}

	a.groups, err = a.buildGroup(ctx)
	if err != nil {
		err = fmt.Errorf("build groups: %w", err)
		return
	}

	a.motionSensors, err = a.buildMotionSensor(ctx)
	if err != nil {
		err = fmt.Errorf("build motion sensor: %w", err)
		return
	}

	return nil
}
