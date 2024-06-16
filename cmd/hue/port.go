package main

import (
	"net/http"
)

func newPort(service service) http.Handler {
	mux := http.NewServeMux()

	mux.Handle("/api/{resource}/{id}", service.hue.Handler())

	service.renderer.Register(mux, service.hue.TemplateFunc)

	return mux
}
