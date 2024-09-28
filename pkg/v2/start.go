package v2

import (
	"context"
	"fmt"
	"log/slog"
	"sort"
	"strings"
)

func (s *Service) Start(ctx context.Context) {
	s.streamIndefinitely(ctx.Done())
}

func (s *Service) Init(ctx context.Context) (err error) {
	slog.Info("Initializing V2...")
	defer slog.Info("Initialization V2 done.")

	var tapDevices []Device
	var motionDevices []Device

	devicePowers, err := list[DevicePower](ctx, s.req, "device_power")
	if err != nil {
		return fmt.Errorf("list devices' powers: %w", err)
	}

	sort.Sort(DevicePowerByOwner(devicePowers))

	if err = s.streamDevices(ctx, func(device Device) {
		tap := strings.EqualFold(device.ProductData.ProductName, "Hue tap switch")
		dial := strings.EqualFold(device.ProductData.ProductName, "Hue tap dial switch")

		if tap || dial {
			tapDevices = append(tapDevices, device)
		}
	}, func(device Device) {
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

	s.motionSensors, err = s.buildMotionSensor(ctx, motionDevices, devicePowers)
	if err != nil {
		err = fmt.Errorf("build motion sensor: %w", err)
		return
	}

	s.taps, err = s.buildTaps(tapDevices, devicePowers)
	if err != nil {
		err = fmt.Errorf("build motion sensor: %w", err)
		return
	}

	return nil
}
