package iot

import (
	"flag"
	"fmt"
	"html/template"
	"net/http"
	"path"
	"strings"
	"sync"

	"github.com/ViBiOh/httputils/v3/pkg/flags"
	"github.com/ViBiOh/httputils/v3/pkg/httperror"
	"github.com/ViBiOh/httputils/v3/pkg/logger"
	"github.com/ViBiOh/httputils/v3/pkg/templates"
	"github.com/ViBiOh/iot/pkg/mqtt"
	"github.com/ViBiOh/iot/pkg/provider"
	"github.com/ViBiOh/iot/pkg/sha"
)

const (
	maxAllowedErrors = 5
	hoursInDay       = 24
	minutesInHours   = 60
	iotSource        = "iot"

	svgPath = "/svg"
)

var (
	hours   []string
	minutes []string
)

func init() {
	hours = make([]string, hoursInDay)
	for i := 0; i < hoursInDay; i++ {
		hours[i] = fmt.Sprintf("%02d", i)
	}

	minutes = make([]string, minutesInHours)
	for i := 0; i < minutesInHours; i++ {
		minutes[i] = fmt.Sprintf("%02d", i)
	}
}

// Config of package
type Config struct {
	AssetsDirectory *string
	subscribe       *string
	publish         *string
	prometheus      *bool
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
	prometheus     bool
}

// Flags adds flags for configuring package
func Flags(fs *flag.FlagSet, prefix string) Config {
	return Config{
		AssetsDirectory: flags.New(prefix, "iot").Name("AssetsDirectory").Default("").Label("Assets directory (static and templates)").ToString(fs),
		subscribe:       flags.New(prefix, "iot").Name("Subscribe").Default("").Label("Topic to subscribe to").ToString(fs),
		publish:         flags.New(prefix, "iot").Name("Publish").Default("worker").Label("Topic to publish to").ToString(fs),
		prometheus:      flags.New(prefix, "iot").Name("Prometheus").Default(false).Label("Expose Prometheus metrics").ToBool(fs),
	}
}

func getTemplate(filesTemplates []string) *template.Template {
	return template.Must(template.New("iot").Funcs(template.FuncMap{
		"sha": sha.Sha1,
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
		"humidity": func(value float32) string {
			switch {
			case value >= 80:
				return "tint?fill=crimson"
			case value >= 60:
				return "tint?fill=darkorange"
			case value >= 40:
				return "tint?fill=limegreen"
			case value >= 20:
				return "tint?fill=darkorange"
			default:
				return "tint?fill=crimson"
			}
		},
	}).ParseFiles(filesTemplates...))
}

// New creates new App from Config
func New(config Config, providers map[string]provider.Provider, mqttClient *mqtt.App) *App {
	filesTemplates, err := templates.GetTemplates(path.Join(*config.AssetsDirectory, "templates"), ".html")
	if err != nil {
		logger.Error("%s", err)
	}

	app := &App{
		tpl:             getTemplate(filesTemplates),
		providers:       providers,
		workerProviders: make(map[string]provider.WorkerProvider, 0),
		workerCalls:     sync.Map{},

		mqttClient:     mqttClient,
		subscribeTopic: strings.TrimSpace(*config.subscribe),
		publishTopic:   strings.TrimSpace(*config.publish),
		prometheus:     *config.prometheus,
	}

	for _, p := range providers {
		if app.prometheus {
			p.EnablePrometheus()
		}

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
		"Message": message,
		"Hours":   hours,
		"Minutes": minutes,
	}

	if message != nil && message.Level == "error" {
		logger.Error("%s", message.Content)
	}

	for name, provider := range a.providers {
		response[name] = provider.GetData()
	}

	w.Header().Set("content-language", "fr")
	if err := templates.WriteHTMLTemplate(a.tpl.Lookup("iot"), w, response, status); err != nil {
		httperror.InternalServerError(w, err)
	}
}

func (a *App) svgHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tpl := a.tpl.Lookup(fmt.Sprintf("svg-%s", strings.Trim(r.URL.Path, "/")))
		if tpl == nil {
			httperror.NotFound(w)
			return
		}

		w.Header().Set("Content-Type", "image/svg+xml")
		if err := tpl.Execute(w, r.URL.Query().Get("fill")); err != nil {
			httperror.InternalServerError(w, err)
		}
	})
}

// Handler create Handler with given App context
func (a *App) Handler() http.Handler {
	strippedSvgHandler := http.StripPrefix(svgPath, a.svgHandler())

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/svg") {
			strippedSvgHandler.ServeHTTP(w, r)
			return
		}

		a.RenderDashboard(w, r, http.StatusOK, nil)
	})
}
