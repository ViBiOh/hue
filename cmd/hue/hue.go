package main

import (
	"context"
	"embed"
	"fmt"
	"net/http"

	_ "net/http/pprof"

	"github.com/ViBiOh/httputils/v4/pkg/alcotest"
	"github.com/ViBiOh/httputils/v4/pkg/cors"
	"github.com/ViBiOh/httputils/v4/pkg/health"
	"github.com/ViBiOh/httputils/v4/pkg/httputils"
	"github.com/ViBiOh/httputils/v4/pkg/logger"
	"github.com/ViBiOh/httputils/v4/pkg/owasp"
	"github.com/ViBiOh/httputils/v4/pkg/recoverer"
	"github.com/ViBiOh/httputils/v4/pkg/renderer"
	"github.com/ViBiOh/httputils/v4/pkg/request"
	"github.com/ViBiOh/httputils/v4/pkg/server"
	"github.com/ViBiOh/httputils/v4/pkg/telemetry"
	"github.com/ViBiOh/hue/pkg/hue"
	v2 "github.com/ViBiOh/hue/pkg/v2"
)

//go:embed templates static
var content embed.FS

func main() {
	ctx := context.Background()

	config, err := newConfig()
	logger.FatalfOnErr(ctx, err, "config")

	alcotest.DoAndExit(config.alcotest)

	logger.Init(config.logger)

	telemetryService, err := telemetry.New(ctx, config.telemetry)
	logger.FatalfOnErr(ctx, err, "create telemetry")

	defer telemetryService.Close(ctx)

	logger.AddOpenTelemetryToDefaultLogger(telemetryService)
	request.AddOpenTelemetryToDefaultClient(telemetryService.MeterProvider(), telemetryService.TracerProvider())

	go func() {
		fmt.Println(http.ListenAndServe("localhost:9999", http.DefaultServeMux))
	}()

	appServer := server.New(config.http)
	healthService := health.New(ctx, config.health)

	rendererService, err := renderer.New(config.renderer, content, hue.FuncMap, telemetryService.MeterProvider(), telemetryService.TracerProvider())
	logger.FatalfOnErr(ctx, err, "create renderer")

	v2Service, err := v2.New(config.hueV2, telemetryService.MeterProvider())
	logger.FatalfOnErr(ctx, err, "create hue v2")

	hueService, err := hue.New(config.hue, rendererService, v2Service)
	logger.FatalfOnErr(ctx, err, "create hue")

	doneCtx := healthService.DoneCtx()
	endCtx := healthService.EndCtx()

	err = v2Service.Init(doneCtx)
	logger.FatalfOnErr(ctx, err, "init v2")

	go hueService.Start(doneCtx)
	go v2Service.Start(doneCtx)

	go appServer.Start(endCtx, httputils.Handler(newPort(hueService, rendererService), healthService, recoverer.Middleware, telemetryService.Middleware("http"), owasp.New(config.owasp).Middleware, cors.New(config.cors).Middleware))

	healthService.WaitForTermination(appServer.Done())
	server.GracefulWait(appServer.Done())
}
