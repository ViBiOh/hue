package hue

type configHue struct {
	Schedules     []ScheduleConfig
	Sensors       []configSensor
	Taps          []configTap
	MotionSensors motionSensors `json:"motion_sensors"`
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
	Rotary  bool
}

type configTapButton struct {
	ID     string
	State  string
	Groups []string
	Lights []string
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
