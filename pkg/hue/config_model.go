package hue

import v2 "github.com/ViBiOh/hue/pkg/v2"

type configHue struct {
	Schedules     []ScheduleConfig
	Sensors       []configSensor
	Taps          []configTap
	MotionSensors motionSensors `json:"motion_sensors"`
	Webhooks      []v2.Webhooks `json:"webhooks"`
}

type configSensor struct {
	ID            string
	LightSensorID string
	CompanionID   string
	OffDelay      string
	Groups        []string
}

type configTap struct {
	ID      string
	Buttons []configTapButton
}

type configTapButton struct {
	ID     string
	State  string
	Groups []string
	Rule   Rule
}

type motionSensors struct {
	Crons []motionSensorCron `json:"crons"`
}

type motionSensorCron struct {
	Hour     string   `json:"hour"`
	Timezone string   `json:"timezone"`
	Names    []string `json:"names"`
	Enabled  bool     `json:"enabled"`
}
