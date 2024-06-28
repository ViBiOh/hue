package main

import (
	"net/http"

	"github.com/ViBiOh/httputils/v4/pkg/httputils"
)

func newPort(clients clients, services services) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("POST /api/groups/{id...}", services.hue.HandleGroup)
	mux.HandleFunc("POST /api/schedules/{id...}", services.hue.HandleSchedule)
	mux.HandleFunc("POST /api/sensors/{id...}", services.hue.HandleSensors)

	services.renderer.RegisterMux(mux, services.hue.TemplateFunc)

	return httputils.Handler(mux, clients.health,
		clients.telemetry.Middleware("http"),
		services.owasp.Middleware,
		services.cors.Middleware,
	)
}
