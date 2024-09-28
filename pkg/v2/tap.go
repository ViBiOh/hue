package v2

import (
	"strconv"
	"strings"
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

func (s *Service) handleTapDevice(device Device) {
	if dial := strings.EqualFold(device.ProductData.ProductName, "Hue tap dial switch"); strings.EqualFold(device.ProductData.ProductName, "Hue tap switch") || dial {
		if dial {
			// ugly hack for now, let's come back latter
			id, _ := strconv.Atoi(device.IDV1)
			device.IDV1 = strconv.Itoa(id + 1)
		}

		s.taps[device.ID] = Tap{
			ID:   device.ID,
			IDV1: device.IDV1,
			Name: device.Metadata.Name,
			Dial: dial,
		}
	}
}
