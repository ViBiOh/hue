package netatmo

// StationData contains data retrieved when getting stations datas
type StationData struct {
	Body struct {
		Devices []struct {
			StationName   string `json:"station_name"`
			DashboardData struct {
				Temperature float32
				Humidity    float32
				Noise       float32
				CO2         float32
			} `json:"dashboard_data"`
			Modules []struct {
				ModuleName    string `json:"module_name"`
				DashboardData struct {
					Temperature float32
					Humidity    float32
				} `json:"dashboard_data"`
			} `json:"modules"`
		} `json:"devices"`
	} `json:"body"`
}

type netatmoError struct {
	Error struct {
		Code    int
		Message string
	}
}

type netatmoToken struct {
	AccessToken string `json:"access_token"`
}
