package v2

import (
	"context"
	"fmt"
	"sort"

	"github.com/ViBiOh/httputils/v4/pkg/breaksync"
	"github.com/ViBiOh/httputils/v4/pkg/concurrent"
)

// MotionSensor description
type MotionSensor struct {
	ID           string  `json:"id"`
	Name         string  `json:"name"`
	BatteryState string  `json:"battery_state"`
	LightLevel   int64   `json:"light_level"`
	Temperature  float64 `json:"temperature"`
	BatteryLevel int     `json:"battery_level"`
	Enabled      bool    `json:"enabled"`
	Motion       bool    `json:"motion"`
}

// LightLevel description
type LightLevel struct {
	Owner struct {
		Rid   string `json:"rid"`
		Rtype string `json:"rtype"`
	} `json:"owner"`
	ID    string `json:"id"`
	Light struct {
		LightLevel      int64 `json:"light_level"`
		LightLevelValid bool  `json:"light_level_valid"`
	} `json:"light"`
	Enabled bool `json:"enabled"`
}

// LightLevelByOwner sort LightLevel by Owner
type LightLevelByOwner []LightLevel

func (a LightLevelByOwner) Len() int      { return len(a) }
func (a LightLevelByOwner) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a LightLevelByOwner) Less(i, j int) bool {
	return a[i].Owner.Rid < a[j].Owner.Rid
}

// Motion description
type Motion struct {
	Owner struct {
		Rid   string `json:"rid"`
		Rtype string `json:"rtype"`
	} `json:"owner"`
	ID     string `json:"id"`
	Motion struct {
		Motion      bool `json:"motion"`
		MotionValid bool `json:"motion_valid"`
	} `json:"motion"`
	Enabled bool `json:"enabled"`
}

// MotionByOwner sort Motion by Owner
type MotionByOwner []Motion

func (a MotionByOwner) Len() int      { return len(a) }
func (a MotionByOwner) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a MotionByOwner) Less(i, j int) bool {
	return a[i].Owner.Rid < a[j].Owner.Rid
}

// Temperature description
type Temperature struct {
	Owner struct {
		Rid   string `json:"rid"`
		Rtype string `json:"rtype"`
	} `json:"owner"`
	ID          string `json:"id"`
	Temperature struct {
		Temperature      float64 `json:"temperature"`
		TemperatureValid bool    `json:"temperature_valid"`
	} `json:"temperature"`
	Enabled bool `json:"enabled"`
}

// TemperatureByOwner sort Temperature by Owner
type TemperatureByOwner []Temperature

func (a TemperatureByOwner) Len() int      { return len(a) }
func (a TemperatureByOwner) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a TemperatureByOwner) Less(i, j int) bool {
	return a[i].Owner.Rid < a[j].Owner.Rid
}

func (a *App) buildMotionSensor(ctx context.Context) (map[string]MotionSensor, error) {
	var devices []Device
	var motions []Motion
	var lightLevels []LightLevel
	var temperatures []Temperature
	var devicePowers []DevicePower

	wg := concurrent.NewFailFast(2)

	wg.Go(func() (err error) {
		devices, err = a.getDevices(ctx, "Hue motion sensor")
		if err != nil {
			return fmt.Errorf("unable to list motion sensors: %s", err)
		}

		sort.Sort(DeviceByID(devices))

		return nil
	})

	wg.Go(func() (err error) {
		motions, err = list[Motion](ctx, a.req, "motion")
		if err != nil {
			return fmt.Errorf("unable to list motions: %s", err)
		}

		sort.Sort(MotionByOwner(motions))

		return nil
	})

	wg.Go(func() (err error) {
		lightLevels, err = list[LightLevel](ctx, a.req, "light_level")
		if err != nil {
			return fmt.Errorf("unable to list light levels: %s", err)
		}

		sort.Sort(LightLevelByOwner(lightLevels))

		return nil
	})

	wg.Go(func() (err error) {
		temperatures, err = list[Temperature](ctx, a.req, "temperature")
		if err != nil {
			return fmt.Errorf("unable to list temperatures: %s", err)
		}

		sort.Sort(TemperatureByOwner(temperatures))

		return nil
	})

	wg.Go(func() (err error) {
		devicePowers, err = list[DevicePower](ctx, a.req, "device_power")
		if err != nil {
			return fmt.Errorf("unable to list devices' powers: %s", err)
		}

		sort.Sort(DevicePowerByOwner(devicePowers))

		return nil
	})

	if err := wg.Wait(); err != nil {
		return nil, fmt.Errorf("unable to fetch motion sensors data: %s", err)
	}

	output := make(map[string]MotionSensor, len(devices))

	return output, breaksync.NewSynchronization().
		AddSources(breaksync.NewSliceSource(devices, func(t Device) string {
			return t.ID
		}, breaksync.NewRupture("id", breaksync.Identity))).
		AddSources(breaksync.NewSliceSource(motions, func(t Motion) string {
			return t.Owner.Rid
		}, nil)).
		AddSources(breaksync.NewSliceSource(lightLevels, func(t LightLevel) string {
			return t.Owner.Rid
		}, nil)).
		AddSources(breaksync.NewSliceSource(temperatures, func(t Temperature) string {
			return t.Owner.Rid
		}, nil)).
		AddSources(breaksync.NewSliceSource(devicePowers, func(t DevicePower) string {
			return t.Owner.Rid
		}, nil)).
		Run(func(syncFlags uint64, values []any) error {
			var sensor MotionSensor

			if syncFlags&1 != 0 {
				return nil
			}

			if syncFlags&1 == 0 {
				device := values[0].(Device)
				sensor.ID = device.ID
				sensor.Name = device.Metadata.Name
			}

			if syncFlags&1<<1 == 0 {
				motion := values[1].(Motion)

				sensor.Enabled = motion.Enabled
				sensor.Motion = motion.Motion.Motion
			}

			if syncFlags&1<<2 == 0 {
				sensor.LightLevel = values[2].(LightLevel).Light.LightLevel
			}

			if syncFlags&1<<3 == 0 {
				sensor.Temperature = values[3].(Temperature).Temperature.Temperature
			}

			if syncFlags&1<<4 == 0 {
				devicePower := values[4].(DevicePower)

				sensor.BatteryLevel = devicePower.PowerState.BatteryLevel
				sensor.BatteryState = devicePower.PowerState.BatteryState
			}

			output[sensor.ID] = sensor
			return nil
		})
}
