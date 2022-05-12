package hue

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

	resp, err := a.v2Req.Path("/eventstream/clip/v2").Accept("text/event-stream").WithClient(createInsecureClient(0)).Send(ctx, nil)
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
				logger.Info("Light %s is %t", data.ID, data.On.On)
			case "temperature":
				logger.Info("Temperature of %s is %f", data.ID, data.Temperature.Temperature)
			case "motion":
				logger.Info("Motion of %s is %t", data.ID, data.Motion.Motion)
			case "light_level":
				logger.Info("Light level of %s is %f", data.ID, data.Light.Level)
			case "grouped_light":
				logger.Info("Group %s is at %f brigtness", data.ID, data.Dimming.Brightness)
			default:
				logger.Info("Unknown event received: `%s`", data.Type)
			}
		}
	}
}
