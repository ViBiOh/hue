package main

import (
	"context"
	"embed"
	"fmt"

	"github.com/ViBiOh/httputils/v4/pkg/cors"
	"github.com/ViBiOh/httputils/v4/pkg/logger"
	"github.com/ViBiOh/httputils/v4/pkg/owasp"
	"github.com/ViBiOh/httputils/v4/pkg/renderer"
	"github.com/ViBiOh/httputils/v4/pkg/server"
	"github.com/ViBiOh/hue/pkg/hue"
	v2 "github.com/ViBiOh/hue/pkg/v2"
)

//go:embed templates static
var content embed.FS

type services struct {
	server   *server.Server
	renderer *renderer.Service
	hue      *hue.Service
	huev2    *v2.Service
	cors     cors.Service
	owasp    owasp.Service
}

func newServices(ctx context.Context, config configuration, clients clients) (services, error) {
	var output services
	var err error

	output.server = server.New(config.server)
	output.owasp = owasp.New(config.owasp)
	output.cors = cors.New(config.cors)

	output.renderer, err = renderer.New(ctx, config.renderer, content, hue.FuncMap, clients.telemetry.MeterProvider(), clients.telemetry.TracerProvider())
	if err != nil {
		return output, fmt.Errorf("renderer: %w", err)
	}

	output.huev2, err = v2.New(config.hueV2, clients.telemetry.MeterProvider())
	if err != nil {
		return output, fmt.Errorf("hue v2: %w", err)
	}

	output.hue, err = hue.New(config.hue, clients.telemetry.TracerProvider(), output.renderer, output.huev2)
	if err != nil {
		return output, fmt.Errorf("hue: %w", err)
	}

	return output, nil
}

func (s services) Start(ctx context.Context) {
	err := s.huev2.Init(ctx)
	logger.FatalfOnErr(ctx, err, "init v2")

	go s.hue.Start(ctx)
	go s.huev2.Start(ctx)
}
