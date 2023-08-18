package v2

import (
	"context"
	"fmt"
	"log/slog"
	"os"
)

// Start worker
func (a *App) Start(ctx context.Context) {
	if err := a.initConfig(ctx); err != nil {
		slog.Error("init", "err", err)
		os.Exit(1)
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
