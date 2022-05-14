package hue

import (
	"context"
	"fmt"
	"strings"
)

// Device description
type Device struct {
	ProductData struct {
		ManufacturerName string `json:"manufacturer_name"`
		ModelID          string `json:"model_id"`
		ProductArchetype string `json:"product_archetype"`
		ProductName      string `json:"product_name"`
		SoftwareVersion  string `json:"software_version"`
		Certified        bool   `json:"certified"`
	} `json:"product_data"`
	Metadata struct {
		Archetype string `json:"archetype"`
		Name      string `json:"name"`
	} `json:"metadata"`
	ID       string `json:"id"`
	IDV1     string `json:"id_v1"`
	Type     string `json:"type"`
	Services []struct {
		Rid   string `json:"rid"`
		Rtype string `json:"rtype"`
	} `json:"services"`
}

// DevicePower description
type DevicePower struct {
	Owner struct {
		Rid   string `json:"rid"`
		Rtype string `json:"rtype"`
	} `json:"owner"`
	ID         string `json:"id"`
	IDV1       string `json:"id_v1"`
	Type       string `json:"type"`
	PowerState struct {
		BatteryState string `json:"battery_state"`
		BatteryLevel int    `json:"battery_level"`
	} `json:"power_state"`
}

func (a *App) getDevices(ctx context.Context, productName string) ([]Device, error) {
	devices, err := listV2[Device](ctx, a.v2Req, "/device")
	if err != nil {
		return nil, fmt.Errorf("unable to fetch: %s", err)
	}

	output := devices[:0]
	for _, device := range devices {
		if strings.EqualFold(device.ProductData.ProductName, productName) {
			output = append(output, device)
		}
	}

	return output, nil
}
