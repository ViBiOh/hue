package v2

import (
	"context"
	"strings"
)

type Tap struct {
	ID           string `json:"id"`
	IDV1         string `json:"id_v1"`
	BatteryState string `json:"battery_state"`
	BatteryLevel int64  `json:"battery_level"`
	Dial         bool   `json:"dial"`
}

func (s *Service) buildTaps(ctx context.Context, input <-chan Device) (map[string]Tap, error) {
	output := make(map[string]Tap)

	for device := range input {
		if strings.EqualFold(device.ProductData.ProductName, "Hue tap switch") {
			output[device.ID] = Tap{
				ID:   device.ID,
				IDV1: device.IDV1,
			}
		}
	}

	return output, nil
}
