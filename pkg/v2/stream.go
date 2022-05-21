package v2

import (
	"bufio"
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"net/http"
	"time"

	"github.com/ViBiOh/httputils/v4/pkg/logger"
	"github.com/ViBiOh/httputils/v4/pkg/request"
)

var dataPrefix []byte = []byte("data: ")

// Event from the server sent event
type Event struct {
	Type string `json:"type"`
	Data []struct {
		Owner struct {
			Rid   string `json:"rid"`
			Rtype string `json:"rtype"`
		} `json:"owner"`
		ID   string `json:"id"`
		Type string `json:"type"`

		Temperature struct {
			Temperature float64 `json:"temperature"`
		} `json:"temperature"`

		Light struct {
			Level int64 `json:"light_level"`
		} `json:"light"`

		Motion struct {
			Motion bool `json:"motion"`
		} `json:"motion"`

		PowerState struct {
			BatteryState string `json:"battery_state"`
			BatteryLevel int    `json:"battery_level"`
		}

		Dimming *Dimming `json:"dimming"`
		On      *On      `json:"on"`
	} `json:"data"`
}

func createInsecureClient(timeout time.Duration) *http.Client {
	client := request.CreateClient(timeout, request.NoRedirection)

	if underlyingTransport, ok := client.Transport.(*http.Transport); ok {
		if underlyingTransport.TLSClientConfig == nil {
			underlyingTransport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
		} else {
			underlyingTransport.TLSClientConfig.InsecureSkipVerify = true
		}
	}

	return client
}

func (a *App) stream(done <-chan struct{}) {
	ctx, cancel := context.WithCancel(context.Background())

	resp, err := a.req.Path("/eventstream/clip/v2").Accept("text/event-stream").WithClient(createInsecureClient(0)).Send(ctx, nil)
	if err != nil {
		logger.Error("unable to open stream: %s", err)
	}

	logger.Info("Streaming events from hub...")
	defer logger.Info("Streaming events ended.")

	go func() {
		<-done
		cancel()
	}()

	var events []Event
	var content []byte

	reader := bufio.NewScanner(resp.Body)
	eventStream := make(chan Event, 4)
	go a.handleStreamEvent(eventStream)

	for reader.Scan() {
		content = reader.Bytes()
		if !bytes.HasPrefix(content, dataPrefix) {
			continue
		}

		content = content[len(dataPrefix):]
		if err := json.Unmarshal(content, &events); err != nil {
			logger.Error("unable to parse event `%s`: %s", content, err)
			continue
		}

		for _, event := range events {
			eventStream <- event
		}
	}

	if closeErr := resp.Body.Close(); closeErr != nil {
		logger.Error("unable to close stream: %s", closeErr)
	}
}

func (a *App) handleStreamEvent(events <-chan Event) {
	for event := range events {
		for _, data := range event.Data {
			switch data.Type {
			case "light":
			case "motion":
				a.updateMotion(data.Owner.Rid, data.Motion.Motion)
			case "light_level":
				a.updateLightLevel(data.Owner.Rid, data.Light.Level)
			case "temperature":
				a.updateTemperature(data.Owner.Rid, data.Temperature.Temperature)
			case "device_power":
				a.updateDevicePower(data.Owner.Rid, data.PowerState.BatteryState, data.PowerState.BatteryLevel)
			case "grouped_light":
				if data.Owner.Rtype != "bridge_home" {
					a.updateGroupedLight(data.ID, data.Dimming, data.On)
				}
			default:
				logger.Info("Unknown event received: `%s`", data.Type)
			}
		}
	}
}

func (a *App) updateMotion(owner string, motion bool) {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	if motionSensor, ok := a.motionSensors[owner]; ok {
		motionSensor.Motion = motion
		logger.Info("Motion %t on %s", motion, motionSensor.Name)

		a.motionSensors[owner] = motionSensor
	} else {
		logger.Warn("unknown motion owner ID `%s`", owner)
	}
}

func (a *App) updateLightLevel(owner string, lightLevel int64) {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	if motionSensor, ok := a.motionSensors[owner]; ok {
		motionSensor.LightLevel = lightLevel
		logger.Info("Light level at %d on %s", lightLevel, motionSensor.Name)

		a.motionSensors[owner] = motionSensor
	} else {
		logger.Warn("unknown light level owner ID `%s`", owner)
	}
}

func (a *App) updateTemperature(owner string, temperature float64) {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	if motionSensor, ok := a.motionSensors[owner]; ok {
		motionSensor.Temperature = temperature
		logger.Info("Temperature at %f on %s", temperature, motionSensor.Name)

		a.motionSensors[owner] = motionSensor

		a.setMetric("temperature", motionSensor.Name, temperature)
	} else {
		logger.Warn("unknown temperature owner ID `%s`", owner)
	}
}

func (a *App) updateDevicePower(owner string, batteryState string, batteryLevel int) {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	if motionSensor, ok := a.motionSensors[owner]; ok {
		motionSensor.BatteryState = batteryState
		motionSensor.BatteryLevel = batteryLevel
		logger.Info("Battery at %d%% on %s", batteryLevel, motionSensor.Name)

		a.motionSensors[owner] = motionSensor
	} else {
		logger.Warn("unknown device power owner ID `%s`", owner)
	}
}

func (a *App) updateGroupedLight(owner string, dimming *Dimming, on *On) {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	if group, ok := a.getGroupOfGroupedLight(owner); ok {
		groupedLight := group.GroupedLights[owner]

		if dimming != nil {
			groupedLight.Dimming.Brightness = dimming.Brightness
			logger.Info("Brightness at %f on %s", dimming.Brightness, group.Name)
		}

		if on != nil {
			groupedLight.On.On = on.On
			logger.Info("On at %t on %s", on.On, group.Name)
		}

		group.GroupedLights[owner] = groupedLight
		a.groups[group.ID] = group
	} else {
		logger.Warn("unknown grouped light ID `%s`", owner)
	}
}
