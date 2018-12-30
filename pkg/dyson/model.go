package dyson

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/ViBiOh/httputils/pkg/errors"
	"github.com/ViBiOh/httputils/pkg/logger"
	"github.com/ViBiOh/iot/pkg/mqtt"
	"github.com/grandcat/zeroconf"
)

const (
	sensorMessage       = `ENVIRONMENTAL-CURRENT-SENSOR-DATA`
	stateMessage        = `CURRENT-STATE`
	currentStateMessage = `REQUEST-CURRENT-STATE`
)

// Data stores data fo hub
type Data struct {
	Devices []*Device
}

// Device for Dyson Link
type Device struct {
	Name             string
	LocalCredentials string
	ProductType      string
	ScaleUnit        string
	Serial           string

	Credentials *Credentials           `json:"-"`
	Service     *zeroconf.ServiceEntry `json:"-"`
	MQTT        *mqtt.App              `json:"-"`
	State       State
}

// Credentials contains device's credential
type Credentials struct {
	Serial       string `json:"serial"`
	PasswordHash string `json:"apPasswordHash"`
}

// State of device
type State struct {
	Temperature    float32
	Humidity       float32
	FanStatus      bool
	FanSpeed       string
	FanRotation    bool
	FanHeating     bool
	FanTemperature float32
}

type message struct {
	Message      string                 `json:"msg"`
	Time         string                 `json:"time,omitempty"`
	Data         map[string]string      `json:"data,omitempty"`
	ProductState map[string]interface{} `json:"product-state,omitempty"`
}

// ConnectToMQTT connect to MQTT of device
func (d *Device) ConnectToMQTT(clientID string) error {
	mqtt, err := mqtt.Connect(d.Service.AddrIPv4[0].String(), d.Credentials.Serial, d.Credentials.PasswordHash, clientID, d.Service.Port, false)
	if err != nil {
		return err
	}

	d.MQTT = mqtt
	d.State = State{}

	return nil
}

// SendCommand send a command to the device
func (d *Device) SendCommand(message []byte) error {
	if d.MQTT == nil {
		return errors.New(`no MQTT configured for device %s`, d.Serial)
	}

	return d.MQTT.Publish(fmt.Sprintf(`%s/%s/command`, d.ProductType, d.Credentials.Serial), message)
}

// SubcribeToStatus subscribe to status update of device
func (d *Device) SubcribeToStatus() error {
	if d.MQTT == nil {
		return errors.New(`no MQTT configured for device %s`, d.Serial)
	}

	return d.MQTT.Subscribe(fmt.Sprintf(`%s/%s/status/current`, d.ProductType, d.Credentials.Serial), func(content []byte) {
		var msg message
		err := json.Unmarshal(content, &msg)
		if err != nil {
			logger.Error(`%+v`, errors.WithStack(err))
			return
		}

		if msg.Message == sensorMessage {
			temperature, err := parseTemperature(msg.Data[`tact`])
			if err != nil {
				logger.Error(`%+v`, err)
				return
			}

			humidity, err := parseHumidity(msg.Data[`hact`])
			if err != nil {
				logger.Error(`%+v`, err)
				return
			}

			d.State.Temperature = temperature
			d.State.Humidity = humidity
		} else if msg.Message == stateMessage {
			d.State.FanStatus = readProductState(msg.ProductState[`fmod`]) == `FAN`

			if d.State.FanStatus {
				d.State.FanSpeed = strings.TrimLeft(readProductState(msg.ProductState[`fnsp`]), `0`)
				d.State.FanRotation = readProductState(msg.ProductState[`oson`]) == `ON`
				d.State.FanHeating = readProductState(msg.ProductState[`hmod`]) == `HEAT`

				temperature, err := parseTemperature(readProductState(msg.ProductState[`hmax`]))
				if err != nil {
					logger.Error(`%+v`, err)
					return
				}
				d.State.FanTemperature = temperature
			}
		}
	})
}

// NewCurrentStateMessage return JSON representation of current state message
func NewCurrentStateMessage() ([]byte, error) {
	msg := message{
		Message: currentStateMessage,
		Time:    time.Now().Format(time.RFC3339),
	}

	payload, err := json.Marshal(msg)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return payload, nil
}
