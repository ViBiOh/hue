package main

import (
	"context"
	"fmt"
	"net/http"

	_ "net/http/pprof"

	"github.com/ViBiOh/httputils/v4/pkg/alcotest"
	"github.com/ViBiOh/httputils/v4/pkg/cors"
	"github.com/ViBiOh/httputils/v4/pkg/httputils"
	"github.com/ViBiOh/httputils/v4/pkg/logger"
	"github.com/ViBiOh/httputils/v4/pkg/owasp"
	"github.com/ViBiOh/httputils/v4/pkg/recoverer"
	"github.com/ViBiOh/httputils/v4/pkg/server"
)

func main() {
	ctx := context.Background()

	config, err := newConfig()
	logger.FatalfOnErr(ctx, err, "config")

	alcotest.DoAndExit(config.alcotest)

	go func() {
		fmt.Println(http.ListenAndServe("localhost:9999", http.DefaultServeMux))
	}()

	client, err := newClient(ctx, config)
	logger.FatalfOnErr(ctx, err, "client")
	defer client.Close(ctx)

	service, err := newService(config, client)
	logger.FatalfOnErr(ctx, err, "client")

	doneCtx := client.health.DoneCtx()

	go service.Start(doneCtx)

	appServer := server.New(config.http)

	go appServer.Start(
		client.health.EndCtx(),
		httputils.Handler(newPort(service), client.health, recoverer.Middleware, client.telemetry.Middleware("http"), owasp.New(config.owasp).Middleware, cors.New(config.cors).Middleware),
	)

	client.health.WaitForTermination(appServer.Done())
	server.GracefulWait(appServer.Done())
}
