package hue

import (
	"context"
	"time"

	"github.com/ViBiOh/httputils/v4/pkg/cron"
	"github.com/ViBiOh/httputils/v4/pkg/logger"
)

func (a *app) Start(done <-chan struct{}) {
	a.initConfig()

	cron.New().Each(time.Minute).Now().OnError(func(err error) {
		logger.Error("%s", err)
	}).Start(a.refreshState, done)
}

func (a *app) initConfig() {
	if a.config == nil {
		logger.Warn("no config init for hue")
		return
	}

	logger.Info("Configuring hue...")
	defer logger.Info("Configuration done.")

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

func (a *app) refreshState(ctx context.Context) error {
	if err := a.syncGroups(); err != nil {
		return err
	}

	if err := a.syncSchedules(); err != nil {
		return err
	}

	if err := a.syncSensors(); err != nil {
		return err
	}

	scenes, err := a.listScenes(ctx)
	if err != nil {
		return err
	}

	a.mutex.Lock()
	a.scenes = scenes
	a.mutex.Unlock()

	go a.updatePrometheusSensors()

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

func (a *app) syncSensors() error {
	sensors, err := a.listSensors(context.Background())
	if err != nil {
		return err
	}

	a.mutex.Lock()
	a.sensors = sensors
	a.mutex.Unlock()

	return nil
}
