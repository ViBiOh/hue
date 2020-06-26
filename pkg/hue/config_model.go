package hue

type configHue struct {
	Schedules []ScheduleConfig
	Sensors   []configSensor
	Taps      []configTap
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
