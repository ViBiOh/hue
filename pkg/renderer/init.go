package renderer

import "fmt"

const (
	hoursInDay     = 24
	minutesInHours = 60
)

var (
	hours   []string
	minutes []string
)

func init() {
	hours = make([]string, hoursInDay)
	for i := 0; i < hoursInDay; i++ {
		hours[i] = fmt.Sprintf("%02d", i)
	}

	minutes = make([]string, minutesInHours)
	for i := 0; i < minutesInHours; i++ {
		minutes[i] = fmt.Sprintf("%02d", i)
	}
}
