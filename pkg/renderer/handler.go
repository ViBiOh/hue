package renderer

import (
	"fmt"
	"html/template"
	"net/http"
	"net/url"
	"strings"

	"github.com/ViBiOh/httputils/v3/pkg/httperror"
	"github.com/ViBiOh/httputils/v3/pkg/query"
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

type app struct {
	tpl *template.Template

	hueApp hue.App
}

// New creates new App from Config
func New(hueApp hue.App) (App, error) {
	filesTemplates, err := templates.GetTemplates("templates", ".html")
	if err != nil {
		return nil, err
	}

	return &app{
		tpl:    getTemplate(filesTemplates),
		hueApp: hueApp,
	}, nil
}

// Handler create Handler with given App context
func (a app) Handler() http.Handler {
	svgHandler := http.StripPrefix(svgPath, a.svgHandler())

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, svgPath) {
			svgHandler.ServeHTTP(w, r)
			return
		}

		if query.IsRoot(r) {
			a.uiHandler(w, r, http.StatusOK, hue.Message{
				Level:   "success",
				Content: r.URL.Query().Get("message"),
			})
			return
		}

		if message, status := a.hueApp.Handle(r); status >= http.StatusBadRequest {
			a.uiHandler(w, r, status, message)
		} else {
			http.Redirect(w, r, fmt.Sprintf("/?message=%s", url.QueryEscape(message.Content)), http.StatusFound)
		}
	})
}

func (a app) uiHandler(w http.ResponseWriter, r *http.Request, status int, message hue.Message) {
	response := map[string]interface{}{
		"Hue": a.hueApp.Data(),
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
