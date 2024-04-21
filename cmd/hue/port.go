package main

import (
	"net/http"

	"github.com/ViBiOh/httputils/v4/pkg/renderer"
	"github.com/ViBiOh/hue/pkg/hue"
)

func newPort(hue *hue.Service, renderer *renderer.Service) http.Handler {
	mux := http.NewServeMux()

	mux.Handle("/api/{resource}/{id}", hue.Handler())
	mux.Handle("/", renderer.Handler(hue.TemplateFunc))

	return mux
}
