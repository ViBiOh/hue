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

	service, err := newService(ctx, config, clients)
	logger.FatalfOnErr(ctx, err, "client")

	go service.Start(clients.health.DoneCtx())

	appServer := server.New(config.http)

	go appServer.Start(
		clients.health.EndCtx(),
		httputils.Handler(newPort(service), clients.health, clients.telemetry.Middleware("http"), owasp.New(config.owasp).Middleware, cors.New(config.cors).Middleware),
	)

	clients.health.WaitForTermination(appServer.Done())
	server.GracefulWait(appServer.Done())
}
