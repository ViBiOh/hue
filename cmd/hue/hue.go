package main

import (
	"context"

	"github.com/ViBiOh/httputils/v4/pkg/alcotest"
	"github.com/ViBiOh/httputils/v4/pkg/cors"
	"github.com/ViBiOh/httputils/v4/pkg/httputils"
	"github.com/ViBiOh/httputils/v4/pkg/logger"
	"github.com/ViBiOh/httputils/v4/pkg/owasp"
	"github.com/ViBiOh/httputils/v4/pkg/server"
)

func main() {
	config := newConfig()
	alcotest.DoAndExit(config.alcotest)

	ctx := context.Background()

	clients, err := newClient(ctx, config)
	logger.FatalfOnErr(ctx, err, "client")

	defer clients.Close(ctx)
	go clients.Start()

	services, err := newServices(ctx, config, clients)
	logger.FatalfOnErr(ctx, err, "client")

	go services.Start(clients.health.DoneCtx())

	port := newPort(config, services)

	go services.server.Start(
		clients.health.EndCtx(),
		httputils.Handler(port, clients.health, clients.telemetry.Middleware("http"), owasp.New(config.owasp).Middleware, cors.New(config.cors).Middleware),
	)

	clients.health.WaitForTermination(services.server.Done())
	server.GracefulWait(services.server.Done())
}
