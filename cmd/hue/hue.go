package main

import (
	"context"
	"embed"
	"flag"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"

	_ "net/http/pprof"

	"github.com/ViBiOh/flags"
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
	fs := flag.NewFlagSet("hue", flag.ExitOnError)
	fs.Usage = flags.Usage(fs)

	appServerConfig := server.Flags(fs, "")
	healthConfig := health.Flags(fs, "")

	alcotestConfig := alcotest.Flags(fs, "")
	loggerConfig := logger.Flags(fs, "logger")
	telemetryConfig := telemetry.Flags(fs, "telemetry")
	owaspConfig := owasp.Flags(fs, "", flags.NewOverride("Csp", "default-src 'self'; script-src 'httputils-nonce'; style-src 'httputils-nonce'"))
	corsConfig := cors.Flags(fs, "cors")
	rendererConfig := renderer.Flags(fs, "", flags.NewOverride("Title", "Hue"), flags.NewOverride("PublicURL", "https://hue.vibioh.fr"), flags.NewOverride("Templates", nil))

	hueConfig := hue.Flags(fs, "")
	v2Config := v2.Flags(fs, "v2")

	if err := fs.Parse(os.Args[1:]); err != nil {
		log.Fatal(err)
	}

	alcotest.DoAndExit(alcotestConfig)

	logger.Init(loggerConfig)

	ctx := context.Background()

	telemetryService, err := telemetry.New(ctx, telemetryConfig)
	logger.FatalfOnErr(ctx, err, "create telemetry")

	defer telemetryService.Close(ctx)

	logger.AddOpenTelemetryToDefaultLogger(telemetryService)
	request.AddOpenTelemetryToDefaultClient(telemetryService.MeterProvider(), telemetryService.TracerProvider())

	go func() {
		fmt.Println(http.ListenAndServe("localhost:9999", http.DefaultServeMux))
	}()

	appServer := server.New(appServerConfig)
	healthService := health.New(ctx, healthConfig)

	rendererService, err := renderer.New(rendererConfig, content, hue.FuncMap, telemetryService.MeterProvider(), telemetryService.TracerProvider())
	logger.FatalfOnErr(ctx, err, "create renderer")

	v2Service, err := v2.New(v2Config, telemetryService.MeterProvider())
	logger.FatalfOnErr(ctx, err, "create hue v2")

	hueService, err := hue.New(hueConfig, rendererService, v2Service)
	logger.FatalfOnErr(ctx, err, "create hue")

	rendererHandler := rendererService.Handler(hueService.TemplateFunc)

	doneCtx := healthService.DoneCtx()
	endCtx := healthService.EndCtx()

	if err := v2Service.Init(doneCtx); err != nil {
		slog.LogAttrs(ctx, slog.LevelError, "init v2", slog.Any("error", err))
		os.Exit(1)
	}

	go hueService.Start(doneCtx)
	go v2Service.Start(doneCtx)

	go appServer.Start(endCtx, httputils.Handler(rendererHandler, healthService, recoverer.Middleware, telemetryService.Middleware("http"), owasp.New(owaspConfig).Middleware, cors.New(corsConfig).Middleware))

	healthService.WaitForTermination(appServer.Done())
	server.GracefulWait(appServer.Done())
}
