package hue

import (
	"html/template"

	v2 "github.com/ViBiOh/hue/pkg/v2"
)

// FuncMap for template rendering
var FuncMap = template.FuncMap{
	"battery": func(value int64) string {
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
	"temperature": func(value float64) string {
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
	"groupName": func(groups []v2.Group, id string) string {
		for _, group := range groups {
			if group.IDV1 == id {
				return group.Name
			}
		}

		return ""
	},
	"monday":    func() int { return monday },
	"tuesday":   func() int { return tuesday },
	"wednesday": func() int { return wednesday },
	"thursday":  func() int { return thursday },
	"friday":    func() int { return friday },
	"saturday":  func() int { return saturday },
	"sunday":    func() int { return sunday },
}
