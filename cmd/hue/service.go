package main

import (
	"context"
	"embed"
	"fmt"

	"github.com/ViBiOh/httputils/v4/pkg/logger"
	"github.com/ViBiOh/httputils/v4/pkg/renderer"
	"github.com/ViBiOh/hue/pkg/hue"
	v2 "github.com/ViBiOh/hue/pkg/v2"
)

//go:embed templates static
var content embed.FS

type service struct {
	hue      *hue.Service
	huev2    *v2.Service
	renderer *renderer.Service
}

func newService(config configuration, client client) (service, error) {
	var output service
	var err error

	output.renderer, err = renderer.New(config.renderer, content, hue.FuncMap, client.telemetry.MeterProvider(), client.telemetry.TracerProvider())
	if err != nil {
		return output, fmt.Errorf("renderer: %w", err)
	}

	output.huev2, err = v2.New(config.hueV2, client.telemetry.MeterProvider())
	if err != nil {
		return output, fmt.Errorf("hue v2: %w", err)
	}

	output.hue, err = hue.New(config.hue, output.renderer, output.huev2)
	if err != nil {
		return output, fmt.Errorf("hue: %w", err)
	}

	return output, nil
}

func (s service) Start(ctx context.Context) {
	err := s.huev2.Init(ctx)
	logger.FatalfOnErr(ctx, err, "init v2")

	go s.hue.Start(ctx)

	s.huev2.Start(ctx)
}
