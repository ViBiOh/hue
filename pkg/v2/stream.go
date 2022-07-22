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
		Motion     *MotionValue    `json:"motion,omitempty"`
		Dimming    *Dimming        `json:"dimming,omitempty"`
		On         *On             `json:"on,omitempty"`
		Enabled    *bool           `json:"enabled,omitempty"`
		Owner      deviceReference `json:"owner"`
		ID         string          `json:"id"`
		Type       string          `json:"type"`
		PowerState struct {
			BatteryState string `json:"battery_state"`
			BatteryLevel int64  `json:"battery_level"`
		}
		Light struct {
			Level int64 `json:"light_level"`
		} `json:"light"`
		Temperature struct {
			Temperature float64 `json:"temperature"`
		} `json:"temperature"`
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

func (a *App) streamIndefinitely(done <-chan struct{}) {
	for {
		a.stream(done)

		select {
		case <-done:
			return
		default:
			logger.Warn("Streaming was ended before done receive, restarting in 5sec...")
			time.Sleep(5 * time.Second)
		}
	}
}

func (a *App) stream(done <-chan struct{}) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	resp, err := a.req.Path("/eventstream/clip/v2").Accept("text/event-stream").WithClient(createInsecureClient(0)).Send(ctx, nil)
	if err != nil {
		logger.Error("unable to open stream: %s", err)
	}

	logger.Info("Streaming events from hub...")
	defer logger.Info("Streaming events ended.")

	go func() {
		select {
		case <-ctx.Done():
		case <-done:
			cancel()
		}
	}()

	reader := bufio.NewScanner(resp.Body)

	for reader.Scan() {
		content := reader.Bytes()
		if !bytes.HasPrefix(content, dataPrefix) {
			continue
		}

		var events []Event
		content = content[len(dataPrefix):]

		if err := json.Unmarshal(content, &events); err != nil {
			logger.Error("unable to parse event `%s`: %s", content, err)
			continue
		}

		for _, event := range events {
			a.handleStreamEvent(event)
		}
	}

	if closeErr := resp.Body.Close(); closeErr != nil {
		logger.Error("unable to close stream: %s", closeErr)
	}
}

func (a *App) handleStreamEvent(event Event) {
	for _, data := range event.Data {
		switch data.Type {
		case "button":
		case "zigbee_connectivity":
		case "motion":
			a.updateMotion(data.Owner.Rid, data.Enabled, data.Motion)
		case "light_level":
			a.updateLightLevel(data.Owner.Rid, data.Light.Level)
		case "temperature":
			a.updateTemperature(data.Owner.Rid, data.Temperature.Temperature)
		case "device_power":
			a.updateDevicePower(data.Owner.Rid, data.PowerState.BatteryState, data.PowerState.BatteryLevel)
		case "light":
			a.updateLight(data.ID, data.On, data.Dimming)
		case "grouped_light":
			a.updateGroupedLight(data.ID, data.On, data.Dimming)
		default:
			logger.Info("unhandled event received: `%s`", data.Type)
		}
	}
}

func (a *App) updateMotion(owner string, enabled *bool, motion *MotionValue) {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	if motionSensor, ok := a.motionSensors[owner]; ok {
		if enabled != nil {
			motionSensor.Enabled = *enabled
			logger.Debug("Motion enabled %t on motion sensor `%s`", motionSensor.Enabled, motionSensor.Name)
		}

		if motion != nil {
			motionSensor.Motion = motion.Motion
			logger.Debug("Motion %t on motion sensor `%s`", motionSensor.Motion, motionSensor.Name)
		}

		a.motionSensors[owner] = motionSensor
		if motion.Motion {
			a.setMetric("motion", motionSensor.Name, 1)
		} else {
			a.setMetric("motion", motionSensor.Name, 0)
		}

	} else {
		logger.Warn("unknown motion owner ID `%s`", owner)
	}
}

func (a *App) updateLightLevel(owner string, lightLevel int64) {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	if motionSensor, ok := a.motionSensors[owner]; ok {
		motionSensor.LightLevel = lightLevel
		logger.Debug("Light level at %d on motion sensor `%s`", lightLevel, motionSensor.Name)

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
		logger.Debug("Temperature at %f on motion sensor `%s`", temperature, motionSensor.Name)

		a.motionSensors[owner] = motionSensor

		a.setMetric("temperature", motionSensor.Name, temperature)
	} else {
		logger.Warn("unknown temperature owner ID `%s`", owner)
	}
}

func (a *App) updateDevicePower(owner string, batteryState string, batteryLevel int64) {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	if motionSensor, ok := a.motionSensors[owner]; ok {
		motionSensor.BatteryState = batteryState
		motionSensor.BatteryLevel = batteryLevel
		logger.Debug("Battery at %d%% on motion sensor `%s`", batteryLevel, motionSensor.Name)

		a.motionSensors[owner] = motionSensor
	} else {
		logger.Warn("unknown device power owner ID `%s`", owner)
	}
}

func (a *App) updateLight(owner string, on *On, dimming *Dimming) {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	if light, ok := a.lights[owner]; ok {
		if dimming != nil {
			light.Dimming.Brightness = dimming.Brightness
			logger.Debug("Brightness at %f on light `%s`", dimming.Brightness, light.Metadata.Name)
		}

		if on != nil {
			light.On.On = on.On
			logger.Debug("On at %t on light `%s`", on.On, light.Metadata.Name)
		}
	} else {
		logger.Warn("unknown light ID `%s`", owner)
	}
}

func (a *App) updateGroupedLight(owner string, on *On, dimming *Dimming) {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	if group, ok := a.getGroupOfGroupedLight(owner); ok {
		groupedLight := group.GroupedLights[owner]

		if dimming != nil {
			groupedLight.Dimming.Brightness = dimming.Brightness
			logger.Debug("Brightness at %f on group `%s`", dimming.Brightness, group.Name)
		}

		if on != nil {
			groupedLight.On.On = on.On
			logger.Debug("On at %t on group `%s`", on.On, group.Name)
		}

		group.GroupedLights[owner] = groupedLight
		a.groups[group.ID] = group
	} else {
		logger.Warn("unknown grouped light ID `%s`", owner)
	}
}
