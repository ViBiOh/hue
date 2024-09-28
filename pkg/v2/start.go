package v2

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
)

func (s *Service) Start(ctx context.Context) {
	s.streamIndefinitely(ctx.Done())
}

func (s *Service) Init(ctx context.Context) (err error) {
	slog.Info("Initializing V2...")
	defer slog.Info("Initialization V2 done.")

	s.taps = make(map[string]Tap)
	var motionDevices []Device

	if err = s.streamDevices(ctx, s.handleTapDevice, func(device Device) {
		if strings.EqualFold(device.ProductData.ProductName, "Hue motion sensor") {
			motionDevices = append(motionDevices, device)
		}
	}); err != nil {
		err = fmt.Errorf("stream devices: %w", err)
		return
	}

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

	s.motionSensors, err = s.buildMotionSensor(ctx, motionDevices)
	if err != nil {
		err = fmt.Errorf("build motion sensor: %w", err)
		return
	}

	return nil
}
