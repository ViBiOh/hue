package main

import (
	"net/http"
)

func newPort(config configuration, service services) http.Handler {
	mux := http.NewServeMux()

	mux.Handle("/api/{resource}/{id}", service.hue.Handler())
	mux.Handle(config.renderer.PathPrefix+"/", service.renderer.NewServeMux(service.hue.TemplateFunc))

	return mux
}
