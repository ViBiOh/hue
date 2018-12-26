package iot

import (
	"flag"
	"fmt"
	"html/template"
	"net/http"
	"path"
	"strings"
	"sync"

	"github.com/ViBiOh/httputils/pkg/errors"
	"github.com/ViBiOh/httputils/pkg/httperror"
	"github.com/ViBiOh/httputils/pkg/logger"
	"github.com/ViBiOh/httputils/pkg/templates"
	"github.com/ViBiOh/httputils/pkg/tools"
	"github.com/ViBiOh/iot/pkg/mqtt"
	"github.com/ViBiOh/iot/pkg/provider"
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

// Config of package
type Config struct {
	AssetsDirectory *string
	subscribe       *string
	publish         *string
}

// App of package
type App struct {
	tpl             *template.Template
	providers       map[string]provider.Provider
	workerProviders map[string]provider.WorkerProvider
	workerCalls     sync.Map

	mqttClient     *mqtt.App
	subscribeTopic string
	publishTopic   string
}

// Flags adds flags for configuring package
func Flags(fs *flag.FlagSet, prefix string) Config {
	return Config{
		AssetsDirectory: fs.String(tools.ToCamel(fmt.Sprintf(`%sAssetsDirectory`, prefix)), ``, `[iot] Assets directory (static and templates)`),
		subscribe:       fs.String(tools.ToCamel(fmt.Sprintf(`%sSubscribe`, prefix)), ``, `[iot] Topic to subscribe to`),
		publish:         fs.String(tools.ToCamel(fmt.Sprintf(`%sPublish`, prefix)), `worker`, `[iot] Topic to publish to`),
	}
}

// New creates new App from Config
func New(config Config, providers map[string]provider.Provider, mqttClient *mqtt.App) *App {
	filesTemplates, err := templates.GetTemplates(path.Join(*config.AssetsDirectory, `templates`), `.html`)
	if err != nil {
		logger.Error(`%+v`, errors.WithStack(err))
	}

	app := &App{
		tpl: template.Must(template.New(`iot`).Funcs(template.FuncMap{
			`sha`: tools.Sha1,
		}).ParseFiles(filesTemplates...)),
		providers:       providers,
		workerProviders: make(map[string]provider.WorkerProvider, 0),
		workerCalls:     sync.Map{},

		mqttClient:     mqttClient,
		subscribeTopic: strings.TrimSpace(*config.subscribe),
		publishTopic:   strings.TrimSpace(*config.publish),
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

// RenderDashboard render dashboard
func (a *App) RenderDashboard(w http.ResponseWriter, r *http.Request, status int, message *provider.Message) {
	response := map[string]interface{}{
		`Message`: message,
		`Hours`:   hours,
		`Minutes`: minutes,
	}

	if message != nil && message.Level == `error` {
		logger.Error(`%+v`, message.Content)
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
