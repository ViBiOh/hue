package main

import (
	"net/http"
)

func newPort(config configuration, service services) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("POST /api/groups/{id...}", service.hue.HandleGroup)
	mux.HandleFunc("POST /api/schedules/{id...}", service.hue.HandleSchedule)
	mux.HandleFunc("POST /api/sensors/{id...}", service.hue.HandleSensors)
	mux.Handle(config.renderer.PathPrefix+"/", service.renderer.NewServeMux(service.hue.TemplateFunc))

	return mux
}
