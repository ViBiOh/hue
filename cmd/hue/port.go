package main

import (
	"net/http"
)

func newPort(service service) http.Handler {
	mux := http.NewServeMux()

	mux.Handle("/api/{resource}/{id}", service.hue.Handler())
	mux.Handle("/", service.renderer.Handler(service.hue.TemplateFunc))

	return mux
}
