package v2

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
	ID       string            `json:"id"`
	IDV1     string            `json:"id_v1"`
	Type     string            `json:"type"`
	Services []deviceReference `json:"services"`
}

type deviceReference struct {
	Rid   string `json:"rid"`
	Rtype string `json:"rtype"`
}

// DeviceByID sort Device by ID
type DeviceByID []Device

func (a DeviceByID) Len() int      { return len(a) }
func (a DeviceByID) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a DeviceByID) Less(i, j int) bool {
	return a[i].ID < a[j].ID
}

// DevicePower description
type DevicePower struct {
	Owner      deviceReference `json:"owner"`
	ID         string          `json:"id"`
	PowerState struct {
		BatteryState string `json:"battery_state"`
		BatteryLevel int64  `json:"battery_level"`
	} `json:"power_state"`
}

// DevicePowerByOwner sort DevicePower by Owner
type DevicePowerByOwner []DevicePower

func (a DevicePowerByOwner) Len() int      { return len(a) }
func (a DevicePowerByOwner) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a DevicePowerByOwner) Less(i, j int) bool {
	return a[i].Owner.Rid < a[j].Owner.Rid
}

func (a *App) getDevices(ctx context.Context, productName string) ([]Device, error) {
	devices, err := list[Device](ctx, a.req, "device")
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
