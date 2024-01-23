package hue

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/ViBiOh/httputils/v4/pkg/cron"
)

var logError = func(ctx context.Context, err error) {
	slog.LogAttrs(ctx, slog.LevelError, "error", slog.Any("error", err))
}

func (s *Service) Start(ctx context.Context) {
	config := s.initConfig(ctx)

	for _, motionSensorCron := range config.MotionSensors.Crons {
		item := motionSensorCron

		go cron.New().Days().At(item.Hour).In(item.Timezone).OnError(logError).Start(ctx, func(ctx context.Context) error {
			return s.updateSensors(ctx, item.Names, item.Enabled)
		})
	}

	cron.New().Each(time.Minute).Now().OnError(logError).Start(ctx, s.syncSchedules)
}

func (s *Service) initConfig(ctx context.Context) (config configHue) {
	if len(s.configFileName) == 0 {
		slog.WarnContext(ctx, "no config init for hue")
		return
	}

	configFile, err := os.Open(s.configFileName)
	if err != nil {
		slog.LogAttrs(ctx, slog.LevelError, "open config file", slog.Any("error", err))
		return
	}

	if err := json.NewDecoder(configFile).Decode(&config); err != nil {
		slog.LogAttrs(ctx, slog.LevelError, "decode config file", slog.Any("error", err))
		return
	}

	if s.update {
		slog.InfoContext(ctx, "Configuring hue...")
		defer slog.InfoContext(ctx, "Configuration done.")

		if err := s.cleanSchedules(ctx); err != nil {
			slog.LogAttrs(ctx, slog.LevelError, "clean schedule", slog.Any("error", err))
		}

		if err := s.cleanRules(ctx); err != nil {
			slog.LogAttrs(ctx, slog.LevelError, "clean rule", slog.Any("error", err))
		}

		if err := s.cleanScenes(ctx); err != nil {
			slog.LogAttrs(ctx, slog.LevelError, "clean scene", slog.Any("error", err))
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
