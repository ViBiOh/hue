package v2

import (
	"context"
	"fmt"
	"net/http"
	"sort"

	"github.com/ViBiOh/httputils/v4/pkg/breaksync"
	"github.com/ViBiOh/httputils/v4/pkg/concurrent"
)

type MotionSensor struct {
	ID           string  `json:"id"`
	MotionID     string  `json:"motion_id"`
	Name         string  `json:"name"`
	BatteryState string  `json:"battery_state"`
	LightLevel   int64   `json:"light_level"`
	Temperature  float64 `json:"temperature"`
	BatteryLevel int64   `json:"battery_level"`
	Enabled      bool    `json:"enabled"`
	Motion       bool    `json:"motion"`
}

type MotionSensorByName []MotionSensor

func (a MotionSensorByName) Len() int      { return len(a) }
func (a MotionSensorByName) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a MotionSensorByName) Less(i, j int) bool {
	return a[i].Name < a[j].Name
}

type LightLevel struct {
	Owner deviceReference `json:"owner"`
	ID    string          `json:"id"`
	Light struct {
		LightLevel      int64 `json:"light_level"`
		LightLevelValid bool  `json:"light_level_valid"`
	} `json:"light"`
	Enabled bool `json:"enabled"`
}

type LightLevelByOwner []LightLevel

func (a LightLevelByOwner) Len() int      { return len(a) }
func (a LightLevelByOwner) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a LightLevelByOwner) Less(i, j int) bool {
	return a[i].Owner.Rid < a[j].Owner.Rid
}

type MotionValue struct {
	Motion      bool `json:"motion"`
	MotionValid bool `json:"motion_valid"`
}

type ColorTemperature struct {
	Mirek int `json:"mirek"`
}

type Color struct {
	XY struct {
		X float64 `json:"x"`
		Y float64 `json:"y"`
	} `json:"xy"`
}

type Motion struct {
	Owner   deviceReference `json:"owner"`
	ID      string          `json:"id"`
	Motion  MotionValue     `json:"motion"`
	Enabled bool            `json:"enabled"`
}

type MotionByOwner []Motion

func (a MotionByOwner) Len() int      { return len(a) }
func (a MotionByOwner) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a MotionByOwner) Less(i, j int) bool {
	return a[i].Owner.Rid < a[j].Owner.Rid
}

type Temperature struct {
	Owner       deviceReference `json:"owner"`
	ID          string          `json:"id"`
	Temperature struct {
		Temperature      float64 `json:"temperature"`
		TemperatureValid bool    `json:"temperature_valid"`
	} `json:"temperature"`
	Enabled bool `json:"enabled"`
}

type TemperatureByOwner []Temperature

func (a TemperatureByOwner) Len() int      { return len(a) }
func (a TemperatureByOwner) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a TemperatureByOwner) Less(i, j int) bool {
	return a[i].Owner.Rid < a[j].Owner.Rid
}

func (a *App) Sensors() []MotionSensor {
	a.mutex.RLock()
	defer a.mutex.RUnlock()

	output := make([]MotionSensor, len(a.motionSensors))

	i := 0
	for _, item := range a.motionSensors {
		output[i] = item
		i++
	}

	sort.Sort(MotionSensorByName(output))

	return output
}

func (a *App) UpdateSensor(ctx context.Context, id string, enabled bool) (MotionSensor, error) {
	a.mutex.RLock()
	defer a.mutex.RUnlock()

	payload := map[string]interface{}{
		"enabled": enabled,
	}

	motionSensor, ok := a.motionSensors[id]
	if !ok {
		return motionSensor, fmt.Errorf("unknown motion sensor with id `%s`", id)
	}

	_, err := a.req.Method(http.MethodPut).Path("/clip/v2/resource/motion/"+motionSensor.MotionID).JSON(ctx, payload)
	return motionSensor, err
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
			return fmt.Errorf("list motion sensors: %w", err)
		}

		sort.Sort(DeviceByID(devices))

		return nil
	})

	wg.Go(func() (err error) {
		motions, err = list[Motion](ctx, a.req, "motion")
		if err != nil {
			return fmt.Errorf("list motions: %w", err)
		}

		sort.Sort(MotionByOwner(motions))

		return nil
	})

	wg.Go(func() (err error) {
		lightLevels, err = list[LightLevel](ctx, a.req, "light_level")
		if err != nil {
			return fmt.Errorf("list light levels: %w", err)
		}

		sort.Sort(LightLevelByOwner(lightLevels))

		return nil
	})

	wg.Go(func() (err error) {
		temperatures, err = list[Temperature](ctx, a.req, "temperature")
		if err != nil {
			return fmt.Errorf("list temperatures: %w", err)
		}

		sort.Sort(TemperatureByOwner(temperatures))

		return nil
	})

	wg.Go(func() (err error) {
		devicePowers, err = list[DevicePower](ctx, a.req, "device_power")
		if err != nil {
			return fmt.Errorf("list devices' powers: %w", err)
		}

		sort.Sort(DevicePowerByOwner(devicePowers))

		return nil
	})

	if err := wg.Wait(); err != nil {
		return nil, fmt.Errorf("fetch motion sensors data: %w", err)
	}

	output := make(map[string]MotionSensor, len(devices))

	return output, breaksync.NewSynchronization().
		AddSources(breaksync.NewSliceSource(devices, func(t Device) []byte {
			return []byte(t.ID)
		}, breaksync.NewRupture("id", breaksync.RuptureIdentity))).
		AddSources(breaksync.NewSliceSource(motions, func(t Motion) []byte {
			return []byte(t.Owner.Rid)
		}, nil)).
		AddSources(breaksync.NewSliceSource(lightLevels, func(t LightLevel) []byte {
			return []byte(t.Owner.Rid)
		}, nil)).
		AddSources(breaksync.NewSliceSource(temperatures, func(t Temperature) []byte {
			return []byte(t.Owner.Rid)
		}, nil)).
		AddSources(breaksync.NewSliceSource(devicePowers, func(t DevicePower) []byte {
			return []byte(t.Owner.Rid)
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
				sensor.MotionID = motion.ID
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
