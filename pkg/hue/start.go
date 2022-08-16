package hue

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/ViBiOh/httputils/v4/pkg/concurrent"
	"github.com/ViBiOh/httputils/v4/pkg/cron"
	"github.com/ViBiOh/httputils/v4/pkg/logger"
)

var logError = func(err error) {
	logger.Error("%s", err)
}

type syncer func(context.Context) error

// Start worker
func (a *App) Start(done <-chan struct{}) {
	config := a.initConfig()

	for _, motionSensorCron := range config.MotionSensors.Crons {
		item := motionSensorCron

		go cron.New().Days().At(item.Hour).In(item.Timezone).OnError(logError).Start(func(ctx context.Context) error {
			return a.updateSensors(ctx, item.Names, item.Enabled)
		}, done)
	}

	cron.New().Each(time.Minute).Now().OnError(logError).Start(a.refreshState, done)
}

func (a *App) initConfig() (config configHue) {
	if len(a.configFileName) == 0 {
		logger.Warn("no config init for hue")
		return
	}

	configFile, err := os.Open(a.configFileName)
	if err != nil {
		logger.Error("open config file: %s", err)
		return
	}

	if err := json.NewDecoder(configFile).Decode(&config); err != nil {
		logger.Error("decode config file: %s", err)
		return
	}

	if a.update {
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

		a.configureSchedules(ctx, config.Schedules)
		a.configureTap(ctx, config.Taps)
		a.configureMotionSensor(ctx, config.Sensors)
	}

	return
}

func (a *App) updateSensors(ctx context.Context, names []string, enabled bool) error {
	for _, sensor := range a.v2App.Sensors() {
		for _, name := range names {
			if sensor.Name == name {
				if _, err := a.v2App.UpdateSensor(ctx, sensor.ID, enabled); err != nil {
					return fmt.Errorf("update sensor `%s`: %w", sensor.ID, err)
				}
			}
		}
	}

	return nil
}

func (a *App) refreshState(ctx context.Context) error {
	if err := a.syncLights(ctx); err != nil {
		return err
	}

	wg := concurrent.NewLimited(4)

	for _, fn := range a.syncers {
		syncer := fn

		wg.Go(func() {
			if err := syncer(ctx); err != nil {
				logger.Error("error while syncing: %s", err)
			}
		})
	}

	wg.Wait()

	return nil
}

func (a *App) syncLights(ctx context.Context) error {
	lights, err := a.listLights(ctx)
	if err != nil {
		return fmt.Errorf("list lights: %w", err)
	}

	a.mutex.Lock()
	defer a.mutex.Unlock()
	a.lights = lights

	return nil
}

func (a *App) syncGroups(ctx context.Context) error {
	groups, err := a.listGroups(ctx)
	if err != nil {
		return fmt.Errorf("list groups: %w", err)
	}

	a.mutex.Lock()
	defer a.mutex.Unlock()
	a.groups = groups

	return nil
}

func (a *App) syncSchedules(ctx context.Context) error {
	schedules, err := a.listSchedules(ctx)
	if err != nil {
		return fmt.Errorf("list schedules: %w", err)
	}

	a.mutex.Lock()
	defer a.mutex.Unlock()
	a.schedules = schedules

	return nil
}

func (a *App) syncScenes(ctx context.Context) error {
	scenes, err := a.listScenes(ctx)
	if err != nil {
		return fmt.Errorf("list scenes: %w", err)
	}

	a.mutex.Lock()
	defer a.mutex.Unlock()
	a.scenes = scenes

	return nil
}
