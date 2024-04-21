package main

import (
	"flag"
	"os"

	"github.com/ViBiOh/flags"
	"github.com/ViBiOh/httputils/v4/pkg/alcotest"
	"github.com/ViBiOh/httputils/v4/pkg/cors"
	"github.com/ViBiOh/httputils/v4/pkg/health"
	"github.com/ViBiOh/httputils/v4/pkg/logger"
	"github.com/ViBiOh/httputils/v4/pkg/owasp"
	"github.com/ViBiOh/httputils/v4/pkg/renderer"
	"github.com/ViBiOh/httputils/v4/pkg/server"
	"github.com/ViBiOh/httputils/v4/pkg/telemetry"
	"github.com/ViBiOh/hue/pkg/hue"
	v2 "github.com/ViBiOh/hue/pkg/v2"
)

type configuration struct {
	alcotest  *alcotest.Config
	telemetry *telemetry.Config
	logger    *logger.Config
	cors      *cors.Config
	owasp     *owasp.Config
	http      *server.Config
	health    *health.Config

	renderer *renderer.Config
	hue      *hue.Config
	hueV2    *v2.Config
}

func newConfig() (configuration, error) {
	fs := flag.NewFlagSet("hue", flag.ExitOnError)
	fs.Usage = flags.Usage(fs)

	return configuration{
		http:      server.Flags(fs, ""),
		health:    health.Flags(fs, ""),
		alcotest:  alcotest.Flags(fs, ""),
		logger:    logger.Flags(fs, "logger"),
		telemetry: telemetry.Flags(fs, "telemetry"),
		owasp:     owasp.Flags(fs, "", flags.NewOverride("Csp", "default-src 'self'; script-src 'httputils-nonce'; style-src 'httputils-nonce'")),
		cors:      cors.Flags(fs, "cors"),

		renderer: renderer.Flags(fs, "", flags.NewOverride("Title", "Hue"), flags.NewOverride("PublicURL", "https://hue.vibioh.fr")),

		hue:   hue.Flags(fs, ""),
		hueV2: v2.Flags(fs, "v2"),
	}, fs.Parse(os.Args[1:])
}
