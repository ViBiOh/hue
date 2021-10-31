package hue

import "html/template"

// FuncMap for template rendering
var FuncMap = template.FuncMap{
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
			return "battery-empty?fill=salmon"
		}
	},
	"temperature": func(value float32) string {
		switch {
		case value >= 28:
			return "thermometer-full?fill=salmon"
		case value >= 24:
			return "thermometer-three-quarters?fill=darkorange"
		case value >= 18:
			return "thermometer-half?fill=limegreen"
		case value >= 14:
			return "thermometer-half?fill=darkorange"
		case value >= 10:
			return "thermometer-quarter?fill=darkorange"
		case value >= 4:
			return "thermometer-empty?fill=salmon"
		default:
			return "snowflake?fill=cornflowerblue"
		}
	},
	"groupName": func(groups map[string]Group, id string) string {
		if group, ok := groups[id]; ok {
			return group.Name
		}
		return ""
	},
}
