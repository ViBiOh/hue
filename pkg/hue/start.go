package hue

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/ViBiOh/httputils/v4/pkg/concurrent"
	"github.com/ViBiOh/httputils/v4/pkg/cron"
)

var logError = func(err error) {
	slog.Error("error", "err", err)
}

type syncer func(context.Context) error

func (s *Service) Start(ctx context.Context) {
	config := s.initConfig(ctx)

	for _, motionSensorCron := range config.MotionSensors.Crons {
		item := motionSensorCron

		go cron.New().Days().At(item.Hour).In(item.Timezone).OnError(logError).Start(ctx, func(ctx context.Context) error {
			return s.updateSensors(ctx, item.Names, item.Enabled)
		})
	}

	cron.New().Each(time.Minute).Now().OnError(logError).Start(ctx, s.refreshState)
}

func (s *Service) initConfig(ctx context.Context) (config configHue) {
	if len(s.configFileName) == 0 {
		slog.Warn("no config init for hue")
		return
	}

	configFile, err := os.Open(s.configFileName)
	if err != nil {
		slog.Error("open config file", "err", err)
		return
	}

	if err := json.NewDecoder(configFile).Decode(&config); err != nil {
		slog.Error("decode config file", "err", err)
		return
	}

	if s.update {
		slog.Info("Configuring hue...")
		defer slog.Info("Configuration done.")

		if err := s.cleanSchedules(ctx); err != nil {
			slog.Error("clean schedule", "err", err)
		}

		if err := s.cleanRules(ctx); err != nil {
			slog.Error("clean rule", "err", err)
		}

		if err := s.cleanScenes(ctx); err != nil {
			slog.Error("clean scene", "err", err)
		}

		s.configureSchedules(ctx, config.Schedules)
		s.configureTap(ctx, config.Taps)
		s.configureMotionSensor(ctx, config.Sensors)
	}

	return
}

func (s *Service) updateSensors(ctx context.Context, names []string, enabled bool) error {
	for _, sensor := range s.v2Service.Sensors() {
		for _, name := range names {
			if sensor.Name == name {
				if _, err := s.v2Service.UpdateSensor(ctx, sensor.ID, enabled); err != nil {
					return fmt.Errorf("update sensor `%s`: %w", sensor.ID, err)
				}
			}
		}
	}

	return nil
}

func (s *Service) refreshState(ctx context.Context) error {
	if err := s.syncLights(ctx); err != nil {
		return err
	}

	wg := concurrent.NewLimiter(4)

	for _, fn := range s.syncers {
		syncer := fn

		wg.Go(func() {
			if err := syncer(ctx); err != nil {
				slog.Error("sync", "err", err)
			}
		})
	}

	wg.Wait()

	return nil
}

func (s *Service) syncLights(ctx context.Context) error {
	lights, err := s.listLights(ctx)
	if err != nil {
		return fmt.Errorf("list lights: %w", err)
	}

	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.lights = lights

	return nil
}

func (s *Service) syncGroups(ctx context.Context) error {
	groups, err := s.listGroups(ctx)
	if err != nil {
		return fmt.Errorf("list groups: %w", err)
	}

	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.groups = groups

	return nil
}

func (s *Service) syncSchedules(ctx context.Context) error {
	schedules, err := s.listSchedules(ctx)
	if err != nil {
		return fmt.Errorf("list schedules: %w", err)
	}

	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.schedules = schedules

	return nil
}

func (s *Service) syncScenes(ctx context.Context) error {
	scenes, err := s.listScenes(ctx)
	if err != nil {
		return fmt.Errorf("list scenes: %w", err)
	}

	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.scenes = scenes

	return nil
}
