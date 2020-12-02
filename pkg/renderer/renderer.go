package renderer

import (
	"fmt"
	"html/template"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/ViBiOh/httputils/v3/pkg/query"
	"github.com/ViBiOh/httputils/v3/pkg/templates"
	"github.com/ViBiOh/hue/pkg/hue"
	"github.com/ViBiOh/hue/pkg/model"
)

const (
	svgPath = "/svg"
)

// App of package
type App interface {
	Handler() http.Handler
}

type app struct {
	tpl     *template.Template
	version string

	hueApp hue.App
}

// New creates new App from Config
func New(hueApp hue.App) (App, error) {
	tpl := template.New("hue").Funcs(template.FuncMap{
		"battery": func(value uint) string {
			switch {
			case value >= 90:
				return "battery-full?fill=limegreen"
			case value >= 75:
				return "battery-three-quarters?fill=limegreen"
			case value >= 50:
				return "battery-half?fill=darkorange"
			case value >= 25:
				return "battery-quarter?fill=darkorange"
			default:
				return "battery-empty?fill=crimson"
			}
		},
		"temperature": func(value float32) string {
			switch {
			case value >= 28:
				return "thermometer-full?fill=crimson"
			case value >= 24:
				return "thermometer-three-quarters?fill=darkorange"
			case value >= 18:
				return "thermometer-half?fill=limegreen"
			case value >= 14:
				return "thermometer-half?fill=darkorange"
			case value >= 10:
				return "thermometer-quarter?fill=darkorange"
			case value >= 4:
				return "thermometer-empty?fill=crimson"
			default:
				return "snowflake?fill=royalblue"
			}
		},
		"groupName": func(groups map[string]hue.Group, id string) string {
			if group, ok := groups[id]; ok {
				return group.Name
			}
			return ""
		},
	})

	filesTemplates, err := templates.GetTemplates("templates", ".html")
	if err != nil {
		return nil, err
	}

	return &app{
		tpl:     template.Must(tpl.ParseFiles(filesTemplates...)),
		version: os.Getenv("VERSION"),
		hueApp:  hueApp,
	}, nil
}

// Handler for request. Should be use with net/http
func (a app) Handler() http.Handler {
	svgHandler := http.StripPrefix(svgPath, a.svg())

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, svgPath) {
			svgHandler.ServeHTTP(w, r)
			return
		}

		if query.IsRoot(r) {
			a.uiHandler(w, r, http.StatusOK, model.Message{
				Level:   r.URL.Query().Get("messageLevel"),
				Content: r.URL.Query().Get("messageContent"),
			})
			return
		}

		message, status := a.hueApp.Handle(r)
		if status >= http.StatusBadRequest {
			a.uiHandler(w, r, status, message)
			return
		}

		http.Redirect(w, r, fmt.Sprintf("/?messageContent=%s", url.QueryEscape(message.Content)), http.StatusFound)

	})
}
