package v2

import (
	"bufio"
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"github.com/ViBiOh/httputils/v4/pkg/request"
)

var dataPrefix []byte = []byte("data: ")

type Event struct {
	Type string `json:"type"`
	Data []struct {
		Motion           *MotionValue      `json:"motion,omitempty"`
		ColorTemperature *ColorTemperature `json:"color_temperature,omitempty"`
		Color            *Color            `json:"color,omitempty"`
		Dimming          *Dimming          `json:"dimming,omitempty"`
		On               *On               `json:"on,omitempty"`
		Enabled          *bool             `json:"enabled,omitempty"`
		Owner            deviceReference   `json:"owner"`
		ID               string            `json:"id"`
		Type             string            `json:"type"`
		PowerState       struct {
			BatteryState string `json:"battery_state"`
			BatteryLevel int64  `json:"battery_level"`
		} `json:"power_state"`
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

func (s *Service) streamIndefinitely(done <-chan struct{}) {
	for {
		s.stream(done)

		select {
		case <-done:
			return
		default:
			slog.Warn("Streaming was ended before done receive, restarting in 5sec...")
			time.Sleep(5 * time.Second)
		}
	}
}

func (s *Service) stream(done <-chan struct{}) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	resp, err := s.req.Path("/eventstream/clip/v2").Accept("text/event-stream").WithClient(createInsecureClient(0)).Send(ctx, nil)
	if err != nil {
		slog.Error("open stream", "error", err)
	}

	slog.Info("Streaming events from hub...")
	defer slog.Info("Streaming events ended.")

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
			slog.Error("parse event", "error", err, "content", content)
			continue
		}

		for _, event := range events {
			s.handleStreamEvent(event)
		}
	}

	if closeErr := resp.Body.Close(); closeErr != nil {
		slog.Error("close stream", "error", closeErr)
	}
}

func (s *Service) handleStreamEvent(event Event) {
	for _, data := range event.Data {
		switch data.Type {
		case "button":
		case "device":
		case "device_software_update":
		case "entertainment":
		case "geofence_client":
		case "zigbee_connectivity":
		case "zigbee_device_discovery":
		case "motion":
			s.updateMotion(data.Owner.Rid, data.Enabled, data.Motion)
		case "light_level":
			s.updateLightLevel(data.Owner.Rid, data.Light.Level)
		case "temperature":
			s.updateTemperature(data.Owner.Rid, data.Temperature.Temperature)
		case "device_power":
			s.updateDevicePower(data.Owner.Rid, data.PowerState.BatteryState, data.PowerState.BatteryLevel)
		case "light":
			s.updateLight(data.ID, data.On, data.Dimming)
		case "grouped_light":
			s.updateGroupedLight(data.ID, data.On, data.Dimming)
		default:
			slog.Info("unhandled event received", "type", data.Type)
		}
	}
}

func (s *Service) updateMotion(owner string, enabled *bool, motion *MotionValue) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if motionSensor, ok := s.motionSensors[owner]; ok {
		if enabled != nil {
			motionSensor.Enabled = *enabled
			slog.Debug("Motion status", "value", motionSensor.Enabled, "sensor", motionSensor.Name)
		}

		if motion != nil {
			motionSensor.Motion = motion.Motion
			slog.Debug("Motion", "motion", motionSensor.Motion, "sensor", motionSensor.Name)
		}

		s.motionSensors[owner] = motionSensor
	} else {
		slog.Warn("unknown motion owner ID", "owner", owner)
	}
}

func (s *Service) updateLightLevel(owner string, lightLevel int64) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if motionSensor, ok := s.motionSensors[owner]; ok {
		motionSensor.LightLevel = lightLevel
		slog.Debug("Light level", "level", lightLevel, "sensor", motionSensor.Name)

		s.motionSensors[owner] = motionSensor
	} else {
		slog.Warn("unknown light level owner ID", "owner", owner)
	}
}

func (s *Service) updateTemperature(owner string, temperature float64) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if motionSensor, ok := s.motionSensors[owner]; ok {
		motionSensor.Temperature = temperature
		slog.Debug("Temperature", "temperature", temperature, "sensor", motionSensor.Name)

		s.motionSensors[owner] = motionSensor
	} else {
		slog.Warn("unknown temperature owner ID", "owner", owner)
	}
}

func (s *Service) updateDevicePower(owner string, batteryState string, batteryLevel int64) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if motionSensor, ok := s.motionSensors[owner]; ok {
		motionSensor.BatteryState = batteryState
		motionSensor.BatteryLevel = batteryLevel
		slog.Debug("Battery", "battery", batteryLevel, "sensor", motionSensor.Name)

		s.motionSensors[owner] = motionSensor
	} else {
		slog.Warn("unknown device power owner ID", "owner", owner)
	}
}

func (s *Service) updateLight(owner string, on *On, dimming *Dimming) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if light, ok := s.lights[owner]; ok {
		if dimming != nil {
			light.Dimming.Brightness = dimming.Brightness
			slog.Debug("Brightness", "brightness", dimming.Brightness, "name", light.Metadata.Name)
		}

		if on != nil {
			light.On.On = on.On
			slog.Debug("Light status", "on", on.On, "name", light.Metadata.Name)
		}
	} else {
		slog.Warn("unknown light ID", "owner", owner)
	}
}

func (s *Service) updateGroupedLight(owner string, on *On, dimming *Dimming) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if group, ok := s.getGroupOfGroupedLight(owner); ok {
		groupedLight := group.GroupedLights[owner]

		if dimming != nil {
			groupedLight.Dimming.Brightness = dimming.Brightness
			slog.Debug("Brightness", "brightness", dimming.Brightness, "group", group.Name)
		}

		if on != nil {
			groupedLight.On.On = on.On
			slog.Debug("Group status", "on", on.On, "group", group.Name)
		}

		group.GroupedLights[owner] = groupedLight
		s.groups[group.ID] = group
	} else {
		slog.Warn("unknown grouped light ID", "owner", owner)
	}
}
