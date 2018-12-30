package dyson

import (
	"strconv"

	"github.com/ViBiOh/httputils/pkg/errors"
)

func parseTemperature(rawTemperature string) (float32, error) {
	rawKelvin, err := strconv.Atoi(rawTemperature)
	if err != nil {
		return 0.0, errors.WithStack(err)
	}

	return kelvinToCelcius(float32(rawKelvin) / 10), nil
}

func parseHumidity(rawHumidity string) (float32, error) {
	humidity, err := strconv.Atoi(rawHumidity)
	if err != nil {
		return 0.0, errors.WithStack(err)
	}

	return float32(humidity), nil
}

func kelvinToCelcius(kelvin float32) float32 {
	return float32(int((kelvin-273.15)*100) / 100)
}
