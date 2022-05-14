package hue

// Room description
type Room struct {
	Metadata struct {
		Archetype string `json:"archetype"`
		Name      string `json:"name"`
	} `json:"metadata"`
	ID       string `json:"id"`
	IDV1     string `json:"id_v1"`
	Type     string `json:"type"`
	Children []struct {
		Rid   string `json:"rid"`
		Rtype string `json:"rtype"`
	} `json:"children"`
	Services []struct {
		Rid   string `json:"rid"`
		Rtype string `json:"rtype"`
	} `json:"services"`
}

// LightV2 description
type LightV2 struct {
	ColorTemperatureDelta struct{} `json:"color_temperature_delta"`
	DimmingDelta          struct{} `json:"dimming_delta"`
	Owner                 struct {
		Rid   string `json:"rid"`
		Rtype string `json:"rtype"`
	} `json:"owner"`
	Metadata struct {
		Archetype string `json:"archetype"`
		Name      string `json:"name"`
	} `json:"metadata"`
	ID      string `json:"id"`
	Mode    string `json:"mode"`
	IDV1    string `json:"id_v1"`
	Type    string `json:"type"`
	Effects struct {
		EffectValues []string `json:"effect_values"`
		Status       string   `json:"status"`
		StatusValues []string `json:"status_values"`
	} `json:"effects"`
	Dynamics struct {
		Status       string   `json:"status"`
		StatusValues []string `json:"status_values"`
		Speed        int64    `json:"speed"`
		SpeedValid   bool     `json:"speed_valid"`
	} `json:"dynamics"`
	Alert struct {
		ActionValues []string `json:"action_values"`
	} `json:"alert"`
	Color struct {
		GamutType string `json:"gamut_type"`
		Gamut     struct {
			Blue struct {
				X float64 `json:"x"`
				Y float64 `json:"y"`
			} `json:"blue"`
			Green struct {
				X float64 `json:"x"`
				Y float64 `json:"y"`
			} `json:"green"`
			Red struct {
				X float64 `json:"x"`
				Y float64 `json:"y"`
			} `json:"red"`
		} `json:"gamut"`
		Xy struct {
			X float64 `json:"x"`
			Y float64 `json:"y"`
		} `json:"xy"`
	} `json:"color"`
	ColorTemperature struct {
		Mirek       int64 `json:"mirek"`
		MirekSchema struct {
			MirekMaximum int64 `json:"mirek_maximum"`
			MirekMinimum int64 `json:"mirek_minimum"`
		} `json:"mirek_schema"`
		MirekValid bool `json:"mirek_valid"`
	} `json:"color_temperature"`
	Dimming struct {
		Brightness  int64   `json:"brightness"`
		MinDimLevel float64 `json:"min_dim_level"`
	} `json:"dimming"`
	On struct {
		On bool `json:"on"`
	} `json:"on"`
}
