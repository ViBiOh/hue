package iot

import (
	"flag"
	"fmt"
	"html/template"
	"net/http"
	"strings"
	"sync"

	"github.com/ViBiOh/httputils/pkg/httperror"
	"github.com/ViBiOh/httputils/pkg/rollbar"
	"github.com/ViBiOh/httputils/pkg/templates"
	"github.com/ViBiOh/httputils/pkg/tools"
	"github.com/ViBiOh/iot/pkg/provider"
	"github.com/gorilla/websocket"
)

const (
	maxAllowedErrors = 5
	hoursInDay       = 24
	minutesInHours   = 60
	iotSource        = `iot`

	svgPath = `/svg`
)

var (
	hours   []string
	minutes []string
)

// App stores informations and secret of API
type App struct {
	tpl       *template.Template
	providers map[string]provider.Provider
	secretKey string

	wsConn     *websocket.Conn
	wsErrCount uint

	workerProviders map[string]provider.WorkerProvider
	workerCalls     sync.Map
}

func init() {
	hours = make([]string, hoursInDay)
	for i := 0; i < hoursInDay; i++ {
		hours[i] = fmt.Sprintf(`%02d`, i)
	}

	minutes = make([]string, minutesInHours)
	for i := 0; i < minutesInHours; i++ {
		minutes[i] = fmt.Sprintf(`%02d`, i)
	}
}

// NewApp creates new App from dependencies and Flags' config
func NewApp(config map[string]*string, providers map[string]provider.Provider) *App {
	filesTemplates, err := templates.GetTemplates(strings.TrimSpace(*config[`templatesDir`]), `.html`)
	if err != nil {
		rollbar.LogError(`error while getting templates: %v`, err)
	}

	app := &App{
		tpl: template.Must(template.New(`iot`).Funcs(template.FuncMap{
			`sha`: tools.Sha1,
		}).ParseFiles(filesTemplates...)),
		providers: providers,
		secretKey: strings.TrimSpace(*config[`secretKey`]),

		workerProviders: make(map[string]provider.WorkerProvider, 0),
		workerCalls:     sync.Map{},
	}

	for _, p := range providers {
		if hubUser, ok := p.(provider.HubUser); ok {
			hubUser.SetHub(app)
		}

		if worker, ok := p.(provider.WorkerProvider); ok {
			app.registerWorker(worker)
		}
	}

	return app
}

// Flags add flags for given prefix
func Flags(prefix string) map[string]*string {
	return map[string]*string{
		`templatesDir`: flag.String(tools.ToCamel(fmt.Sprintf(`%sTemplates`, prefix)), `./templates/`, `[iot] Templates directory`),
		`secretKey`:    flag.String(tools.ToCamel(fmt.Sprintf(`%sSecretKey`, prefix)), ``, `[iot] Secret Key between worker and API`),
	}
}

// RenderDashboard render dashboard
func (a *App) RenderDashboard(w http.ResponseWriter, r *http.Request, status int, message *provider.Message) {
	response := map[string]interface{}{
		`Online`:  a.wsConn != nil,
		`Error`:   a.wsErrCount >= maxAllowedErrors,
		`Message`: message,
		`Hours`:   hours,
		`Minutes`: minutes,
	}

	if message != nil && message.Level == `error` {
		rollbar.LogError(message.Content)
	}

	for name, provider := range a.providers {
		response[name] = provider.GetData()
	}

	w.Header().Set(`content-language`, `fr`)
	if err := templates.WriteHTMLTemplate(a.tpl.Lookup(`iot`), w, response, status); err != nil {
		httperror.InternalServerError(w, err)
	}
}

func (a *App) svgHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tpl := a.tpl.Lookup(fmt.Sprintf(`svg-%s`, strings.Trim(r.URL.Path, `/`)))
		if tpl == nil {
			httperror.NotFound(w)
			return
		}

		w.Header().Set(`Content-Type`, `image/svg+xml`)
		if err := tpl.Execute(w, r.URL.Query().Get(`fill`)); err != nil {
			httperror.InternalServerError(w, err)
		}
	})
}

// Handler create Handler with given App context
func (a *App) Handler() http.Handler {
	usedSvgHandler := http.StripPrefix(svgPath, a.svgHandler())

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, `/svg`) {
			usedSvgHandler.ServeHTTP(w, r)
			return
		}

		a.RenderDashboard(w, r, http.StatusOK, nil)
	})
}
