package renderer

import (
	"flag"
	"fmt"
	"html/template"
	"net/http"
	"path"
	"strings"

	"github.com/ViBiOh/httputils/v3/pkg/flags"
	"github.com/ViBiOh/httputils/v3/pkg/httperror"
	"github.com/ViBiOh/httputils/v3/pkg/logger"
	"github.com/ViBiOh/httputils/v3/pkg/templates"
	"github.com/ViBiOh/iot/pkg/hue"
)

const (
	svgPath = "/svg"
)

var (
	_ App = app{}
)

// App of package
type App interface {
	Handler() http.Handler
}

// Config of package
type Config struct {
	AssetsDirectory *string
}

type app struct {
	tpl *template.Template

	hueApp hue.App
}

// Flags adds flags for configuring package
func Flags(fs *flag.FlagSet, prefix string) Config {
	return Config{
		AssetsDirectory: flags.New(prefix, "hue").Name("AssetsDirectory").Default("").Label("Assets directory (static and templates)").ToString(fs),
	}
}

// New creates new App from Config
func New(config Config, hueApp hue.App) App {
	filesTemplates, err := templates.GetTemplates(path.Join(*config.AssetsDirectory, "templates"), ".html")
	if err != nil {
		logger.Error("%s", err)
	}

	return &app{
		tpl:    getTemplate(filesTemplates),
		hueApp: hueApp,
	}
}

// Handler create Handler with given App context
func (a app) Handler() http.Handler {
	strippedSvgHandler := http.StripPrefix(svgPath, a.svgHandler())

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, svgPath) {
			strippedSvgHandler.ServeHTTP(w, r)
			return
		}

		message, status := a.hueApp.Handle(r)
		a.uiHandler(w, r, status, message)
	})
}

func (a app) uiHandler(w http.ResponseWriter, r *http.Request, status int, message hue.Message) {
	response := map[string]interface{}{
		"Hours":   hours,
		"Minutes": minutes,
		"Hue":     a.hueApp.Data(),
	}

	if len(message.Content) > 0 {
		response["Message"] = message
	}

	if err := templates.ResponseHTMLTemplate(a.tpl.Lookup("iot"), w, response, status); err != nil {
		httperror.InternalServerError(w, err)
	}
}

func (a app) svgHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tpl := a.tpl.Lookup(fmt.Sprintf("svg-%s", strings.Trim(r.URL.Path, "/")))
		if tpl == nil {
			httperror.NotFound(w)
			return
		}

		w.Header().Set("Content-Type", "image/svg+xml")
		if err := templates.WriteTemplate(tpl, w, r.URL.Query().Get("fill"), "text/xml"); err != nil {
			httperror.InternalServerError(w, err)
		}
	})
}
