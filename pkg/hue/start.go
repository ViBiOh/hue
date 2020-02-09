package hue

import (
	"context"
	"time"

	"github.com/ViBiOh/httputils/v3/pkg/cron"
	"github.com/ViBiOh/httputils/v3/pkg/logger"
)

func (a *app) Start() {
	a.initConfig()

	cron.New().Each(time.Minute).Now().Start(a.refreshState, func(err error) {
		logger.Error("%s", err)
	})
}

func (a *app) initConfig() {
	if a.config == nil {
		logger.Warn("no config init for hue")
		return
	}

	ctx := context.Background()

	if err := a.cleanSchedules(ctx); err != nil {
		logger.Error("%s", err)
	}

	if err := a.cleanRules(ctx); err != nil {
		logger.Error("%s", err)
	}

	if err := a.cleanScenes(ctx); err != nil {
		logger.Error("%s", err)
	}

	a.configureSchedules(ctx, a.config.Schedules)
	a.configureTap(ctx, a.config.Taps)
	a.configureMotionSensor(ctx, a.config.Sensors)
}

func (a *app) refreshState(_ time.Time) error {
	if err := a.syncGroups(); err != nil {
		return err
	}

	if err := a.syncSchedules(); err != nil {
		return err
	}

	scenes, err := a.listScenes(context.Background())
	if err != nil {
		return err
	}

	sensors, err := a.listSensors(context.Background())
	if err != nil {
		return err
	}

	a.mutex.Lock()
	a.scenes = scenes
	a.sensors = sensors
	a.mutex.Unlock()

	return nil
}

func (a *app) syncGroups() error {
	groups, err := a.listGroups(context.Background())
	if err != nil {
		return err
	}

	a.mutex.Lock()
	a.groups = groups
	a.mutex.Unlock()

	return nil
}

func (a *app) syncSchedules() error {
	schedules, err := a.listSchedules(context.Background())
	if err != nil {
		return err
	}

	a.mutex.Lock()
	a.schedules = schedules
	a.mutex.Unlock()

	return nil
}
