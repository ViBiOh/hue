package netatmo

const (
	// DevicesAction action for listing devices
	DevicesAction = "devices"

	// Source constant for worker message
	Source = "netatmo"
)

// StationsData contains data retrieved when getting stations datas
type StationsData struct {
	Body struct {
		Devices []*Device `json:"devices"`
	} `json:"body"`
}

// Device contains a device data
type Device struct {
	StationName   string        `json:"station_name"`
	DashboardData DashboardData `json:"dashboard_data"`
	Modules       []struct {
		ModuleName    string        `json:"module_name"`
		DashboardData DashboardData `json:"dashboard_data"`
	} `json:"modules"`
}

// DashboardData contains dashboard data
type DashboardData struct {
	Temperature float32
	Humidity    float32
	Noise       float32
	CO2         float32
}

// Error describes error
type Error struct {
	Error struct {
		Code    int
		Message string
	}
}

// Token describes refresh token response
type Token struct {
	AccessToken string `json:"access_token"`
}
