package hue

import (
	"context"
	"fmt"
)

// MotionSensor description
type MotionSensor struct {
	ID           string  `json:"id"`
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
	IDV1  string `json:"id_v1"`
	Type  string `json:"type"`
	Light struct {
		LightLevel      int64 `json:"light_level"`
		LightLevelValid bool  `json:"light_level_valid"`
	} `json:"light"`
	Enabled bool `json:"enabled"`
}

// Motion description
type Motion struct {
	Owner struct {
		Rid   string `json:"rid"`
		Rtype string `json:"rtype"`
	} `json:"owner"`
	ID     string `json:"id"`
	IDV1   string `json:"id_v1"`
	Type   string `json:"type"`
	Motion struct {
		Motion      bool `json:"motion"`
		MotionValid bool `json:"motion_valid"`
	} `json:"motion"`
	Enabled bool `json:"enabled"`
}

// Temperature description
type Temperature struct {
	Owner struct {
		Rid   string `json:"rid"`
		Rtype string `json:"rtype"`
	} `json:"owner"`
	ID          string `json:"id"`
	IDV1        string `json:"id_v1"`
	Type        string `json:"type"`
	Temperature struct {
		Temperature      float64 `json:"temperature"`
		TemperatureValid bool    `json:"temperature_valid"`
	} `json:"temperature"`
	Enabled bool `json:"enabled"`
}

func (a *App) buildMotionSensor(ctx context.Context) (map[string]MotionSensor, error) {
	motionSensors, err := a.getDevices(ctx, "Hue motion sensor")
	if err != nil {
		return nil, fmt.Errorf("unable to get devices: %s", err)
	}

	output := make(map[string]MotionSensor, len(motionSensors))
	for _, motionSensor := range motionSensors {
		output[motionSensor.ID] = MotionSensor{
			ID: motionSensor.ID,
		}
	}

	return output, nil
}
