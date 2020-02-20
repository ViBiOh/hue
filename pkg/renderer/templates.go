package renderer

import (
	"html/template"
)

func getTemplate(filesTemplates []string) *template.Template {
	return template.Must(template.New("iot").Funcs(template.FuncMap{
		"battery": func(value uint) string {
			switch {
			case value >= 90:
				return "battery-full?fill=limegreen"
			case value >= 75:
				return "battery-three-quarters?fill=limegreen"
			case value >= 50:
				return "battery-half?fill=darkorange"
			case value >= 25:
				return "battery-quarter?fill=darkorange"
			default:
				return "battery-empty?fill=crimson"
			}
		},
		"temperature": func(value float32) string {
			switch {
			case value >= 28:
				return "thermometer-full?fill=crimson"
			case value >= 24:
				return "thermometer-three-quarters?fill=darkorange"
			case value >= 18:
				return "thermometer-half?fill=limegreen"
			case value >= 14:
				return "thermometer-half?fill=darkorange"
			case value >= 10:
				return "thermometer-quarter?fill=darkorange"
			case value >= 4:
				return "thermometer-empty?fill=crimson"
			default:
				return "snowflake?fill=royalblue"
			}
		},
		"humidity": func(value float32) string {
			switch {
			case value >= 80:
				return "tint?fill=crimson"
			case value >= 60:
				return "tint?fill=darkorange"
			case value >= 40:
				return "tint?fill=limegreen"
			case value >= 20:
				return "tint?fill=darkorange"
			default:
				return "tint?fill=crimson"
			}
		},
	}).ParseFiles(filesTemplates...))
}