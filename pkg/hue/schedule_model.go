package hue

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"
	"strings"
)

const (
	sunday = 1 << iota
	saturday
	friday
	thursday
	wednesday
	tuesday
	monday

	weekday = monday | tuesday | wednesday | thursday | friday
	weekend = saturday | sunday
	alldays = weekday | weekend
)

// Schedule description
type Schedule struct {
	ID string `json:"id,omitempty"`
	APISchedule
}

// ByScheduleID sort Schedule by id
type ByScheduleID []Schedule

func (a ByScheduleID) Len() int      { return len(a) }
func (a ByScheduleID) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a ByScheduleID) Less(i, j int) bool {
	return a[i].ID < a[j].ID
}

// APISchedule describe schedule as from Hue API
type APISchedule struct {
	Name      string `json:"name,omitempty"`
	Localtime string `json:"localtime,omitempty"`
	Command   Action `json:"command"`
	Status    string `json:"status,omitempty"`
}

// ScheduleConfig configuration (made simple)
type ScheduleConfig struct {
	Name      string
	Localtime string
	Group     string
	State     string
}

func recurrenceStr(recurrence int) string {
	if recurrence == alldays {
		return "All days"
	}
	if recurrence == weekday {
		return "Week days"
	}
	if recurrence == weekend {
		return "Weekend"
	}

	var days []string

	if recurrence&monday != 0 {
		days = append(days, "Mon")
	}
	if recurrence&tuesday != 0 {
		days = append(days, "Tue")
	}
	if recurrence&wednesday != 0 {
		days = append(days, "Wed")
	}
	if recurrence&thursday != 0 {
		days = append(days, "Thu")
	}
	if recurrence&friday != 0 {
		days = append(days, "Fri")
	}
	if recurrence&saturday != 0 {
		days = append(days, "Sat")
	}
	if recurrence&sunday != 0 {
		days = append(days, "Sun")
	}

	return strings.Join(days, ", ")
}

func (s Schedule) HasDay(dayValue int) bool {
	if !strings.HasPrefix(s.Localtime, "W") {
		return false
	}

	recurrence, err := strconv.Atoi(s.Localtime[1:4])
	if err != nil {
		return false
	}

	return recurrence&dayValue != 0
}

func (s Schedule) ScheduleTime() string {
	separator := strings.Index(s.Localtime, "/T")
	if separator == -1 {
		return ""
	}

	return s.Localtime[separator+2 : separator+7]
}

// FormatLocalTime formats local time of schedules to human-readable version
func (s Schedule) FormatLocalTime() string {
	if !strings.HasPrefix(s.Localtime, "W") {
		return s.Localtime
	}

	recurrence, err := strconv.Atoi(s.Localtime[1:4])
	if err != nil {
		slog.LogAttrs(context.Background(), slog.LevelError, "format time", slog.Any("error", err))
		return s.Localtime
	}

	return fmt.Sprintf("%s at %s", recurrenceStr(recurrence), s.Localtime[6:])
}
