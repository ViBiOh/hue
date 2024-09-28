package v2

import (
	"sort"
	"strconv"
	"strings"

	"github.com/ViBiOh/httputils/v4/pkg/breaksync"
)

type Tap struct {
	ID           string
	IDV1         string
	Name         string
	BatteryState string
	BatteryLevel int64
	Dial         bool
}

func (s *Service) Taps() []Tap {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	output := make([]Tap, 0, len(s.taps))

	for _, item := range s.taps {
		output = append(output, item)
	}

	return output
}

func (s *Service) buildTaps(devices []Device, devicePowers []DevicePower) (map[string]Tap, error) {
	sort.Sort(DeviceByID(devices))

	output := make(map[string]Tap, len(devices))

	return output, breaksync.NewSynchronization().
		AddSources(breaksync.NewSliceSource(devices, func(t Device) []byte {
			return []byte(t.ID)
		}, breaksync.NewRupture("id", breaksync.RuptureIdentity))).
		AddSources(breaksync.NewSliceSource(devicePowers, func(t DevicePower) []byte {
			return []byte(t.Owner.Rid)
		}, nil)).
		Run(func(syncFlags uint, values []any) error {
			var tap Tap

			if syncFlags&1 != 0 {
				return nil
			}

			if syncFlags&1 == 0 {
				device := values[0].(Device)

				dial := strings.EqualFold(device.ProductData.ProductName, "Hue tap dial switch")
				if dial {
					// ugly hack for now, let's come back latter
					id, _ := strconv.Atoi(device.IDV1)
					device.IDV1 = strconv.Itoa(id + 1)
				}

				tap.ID = device.ID
				tap.IDV1 = device.IDV1
				tap.Name = device.Metadata.Name
				tap.Dial = dial
			}

			if syncFlags&1<<1 == 0 {
				devicePower := values[1].(DevicePower)

				tap.BatteryLevel = devicePower.PowerState.BatteryLevel
				tap.BatteryState = devicePower.PowerState.BatteryState
			}

			output[tap.ID] = tap
			return nil
		})
}
