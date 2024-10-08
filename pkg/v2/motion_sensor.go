package v2

import (
	"context"
	"fmt"
	"net/http"
	"sort"
	"strings"

	"github.com/ViBiOh/httputils/v4/pkg/breaksync"
	"github.com/ViBiOh/httputils/v4/pkg/concurrent"
)

type MotionSensor struct {
	ID           string `json:"id"`
	IDV1         string `json:"id_v1"`
	MotionID     string `json:"motion_id"`
	Name         string `json:"name"`
	BatteryState string `json:"battery_state"`

	LightLevelID    string `json:"light_level_id"`
	LightLevelIDV1  string `json:"light_level_id_v1"`
	LightLevelValue int64  `json:"light_level"`

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
	IDV1  string          `json:"id_v1"`
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

func (s *Service) Sensors() []MotionSensor {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	output := make([]MotionSensor, 0, len(s.motionSensors))

	for _, item := range s.motionSensors {
		output = append(output, item)
	}

	sort.Sort(MotionSensorByName(output))

	return output
}

func (s *Service) UpdateSensor(ctx context.Context, id string, enabled bool) (MotionSensor, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	payload := map[string]interface{}{
		"enabled": enabled,
	}

	motionSensor, ok := s.motionSensors[id]
	if !ok {
		return motionSensor, fmt.Errorf("unknown motion sensor with id `%s`", id)
	}

	_, err := s.req.Method(http.MethodPut).Path("/clip/v2/resource/motion/"+motionSensor.MotionID).JSON(ctx, payload)
	return motionSensor, err
}

func (s *Service) buildMotionSensor(ctx context.Context, devices []Device, devicePowers []DevicePower) (map[string]MotionSensor, error) {
	var motions []Motion
	var lightLevels []LightLevel
	var temperatures []Temperature

	wg := concurrent.NewFailFast(2)

	wg.Go(func() (err error) {
		motions, err = list[Motion](ctx, s.req, "motion")
		if err != nil {
			return fmt.Errorf("list motions: %w", err)
		}

		sort.Sort(MotionByOwner(motions))

		return nil
	})

	wg.Go(func() (err error) {
		lightLevels, err = list[LightLevel](ctx, s.req, "light_level")
		if err != nil {
			return fmt.Errorf("list light levels: %w", err)
		}

		sort.Sort(LightLevelByOwner(lightLevels))

		return nil
	})

	wg.Go(func() (err error) {
		temperatures, err = list[Temperature](ctx, s.req, "temperature")
		if err != nil {
			return fmt.Errorf("list temperatures: %w", err)
		}

		sort.Sort(TemperatureByOwner(temperatures))

		return nil
	})

	if err := wg.Wait(); err != nil {
		return nil, fmt.Errorf("fetch motion sensors data: %w", err)
	}

	sort.Sort(DeviceByID(devices))

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
		Run(func(syncFlags uint, values []any) error {
			var sensor MotionSensor

			if syncFlags&1 != 0 {
				return nil
			}

			if syncFlags&1 == 0 {
				device := values[0].(Device)
				sensor.ID = device.ID
				sensor.IDV1 = device.IDV1
				sensor.Name = device.Metadata.Name
			}

			if syncFlags&1<<1 == 0 {
				motion := values[1].(Motion)

				sensor.Enabled = motion.Enabled
				sensor.Motion = motion.Motion.Motion
				sensor.MotionID = motion.ID
			}

			if syncFlags&1<<2 == 0 {
				lightLevel := values[2].(LightLevel)

				sensor.LightLevelID = lightLevel.ID
				sensor.LightLevelIDV1 = strings.TrimPrefix(lightLevel.IDV1, "/sensors/")
				sensor.LightLevelValue = lightLevel.Light.LightLevel
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
