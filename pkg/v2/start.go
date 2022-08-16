package v2

import (
	"context"
	"fmt"

	"github.com/ViBiOh/httputils/v4/pkg/logger"
)

// Start worker
func (a *App) Start(done <-chan struct{}) {
	if err := a.initConfig(); err != nil {
		logger.Fatal(err)
	}

	go a.streamIndefinitely(done)
}

func (a *App) initConfig() (err error) {
	ctx := context.Background()

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
