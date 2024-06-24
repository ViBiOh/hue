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
	"github.com/ViBiOh/httputils/v4/pkg/pprof"
	"github.com/ViBiOh/httputils/v4/pkg/renderer"
	"github.com/ViBiOh/httputils/v4/pkg/server"
	"github.com/ViBiOh/httputils/v4/pkg/telemetry"
	"github.com/ViBiOh/hue/pkg/hue"
	v2 "github.com/ViBiOh/hue/pkg/v2"
)

type configuration struct {
	logger    *logger.Config
	alcotest  *alcotest.Config
	telemetry *telemetry.Config
	pprof     *pprof.Config
	health    *health.Config

	server   *server.Config
	owasp    *owasp.Config
	cors     *cors.Config
	renderer *renderer.Config

	hue   *hue.Config
	hueV2 *v2.Config
}

func newConfig() configuration {
	fs := flag.NewFlagSet("hue", flag.ExitOnError)
	fs.Usage = flags.Usage(fs)

	config := configuration{
		logger:    logger.Flags(fs, "logger"),
		alcotest:  alcotest.Flags(fs, ""),
		telemetry: telemetry.Flags(fs, "telemetry"),
		pprof:     pprof.Flags(fs, "pprof"),
		health:    health.Flags(fs, ""),

		server:   server.Flags(fs, ""),
		owasp:    owasp.Flags(fs, "", flags.NewOverride("Csp", "default-src 'self'; script-src 'httputils-nonce'; style-src 'httputils-nonce'")),
		cors:     cors.Flags(fs, "cors"),
		renderer: renderer.Flags(fs, "", flags.NewOverride("Title", "Hue"), flags.NewOverride("PublicURL", "https://hue.vibioh.fr")),

		hue:   hue.Flags(fs, ""),
		hueV2: v2.Flags(fs, "v2"),
	}

	_ = fs.Parse(os.Args[1:])

	return config
}
