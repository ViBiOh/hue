package hue

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/ViBiOh/httputils/v4/pkg/cron"
	"github.com/ViBiOh/httputils/v4/pkg/logger"
)

type syncer func() error

// Start worker
func (a *App) Start(done <-chan struct{}) {
	a.initConfig()

	cron.New().Each(time.Minute).Now().OnError(func(err error) {
		logger.Error("%s", err)
	}).Start(a.refreshState, done)
}

func (a *App) initConfig() {
	if len(a.configFileName) == 0 {
		logger.Warn("no config init for hue")
		return
	}

	configFile, err := os.Open(a.configFileName)
	if err != nil {
		logger.Error("unable to open config file: %s", err)
		return
	}

	var config configHue
	if err := json.NewDecoder(configFile).Decode(&config); err != nil {
		logger.Error("unable to decode config file: %s", err)
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

	a.configureSchedules(ctx, config.Schedules)
	a.configureTap(ctx, config.Taps)
	a.configureMotionSensor(ctx, config.Sensors)
}

func (a *App) refreshState(ctx context.Context) error {
	if err := a.syncLights(); err != nil {
		return err
	}

	var wg sync.WaitGroup
	wg.Add(len(a.syncers))

	for _, fn := range a.syncers {
		go func(syncer syncer) {
			defer wg.Done()
			if err := syncer(); err != nil {
				logger.Error("error while syncing: %s", err)
			}
		}(fn)
	}

	wg.Wait()

	return nil
}

func (a *App) syncLights() error {
	lights, err := a.listLights(context.Background())
	if err != nil {
		return fmt.Errorf("unable to list lights: %s", err)
	}

	a.mutex.Lock()
	defer a.mutex.Unlock()
	a.lights = lights

	return nil
}

func (a *App) syncGroups() error {
	groups, err := a.listGroups(context.Background())
	if err != nil {
		return fmt.Errorf("unable to list groups: %s", err)
	}

	a.mutex.Lock()
	defer a.mutex.Unlock()
	a.groups = groups

	return nil
}

func (a *App) syncSchedules() error {
	schedules, err := a.listSchedules(context.Background())
	if err != nil {
		return fmt.Errorf("unable to list schedules: %s", err)
	}

	a.mutex.Lock()
	defer a.mutex.Unlock()
	a.schedules = schedules

	return nil
}

func (a *App) syncSensors() error {
	sensors, err := a.listSensors(context.Background())
	if err != nil {
		return fmt.Errorf("unable to list sensors: %s", err)
	}

	a.mutex.Lock()
	defer a.mutex.Unlock()
	a.sensors = sensors

	return nil
}

func (a *App) syncScenes() error {
	scenes, err := a.listScenes(context.Background())
	if err != nil {
		return fmt.Errorf("unable to list scenes: %s", err)
	}

	a.mutex.Lock()
	defer a.mutex.Unlock()
	a.scenes = scenes

	return nil
}
