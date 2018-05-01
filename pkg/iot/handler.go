package iot

import (
	"flag"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"sync"

	"github.com/ViBiOh/httputils/pkg/httperror"
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
)

var (
	hours   []string
	minutes []string
)

// App stores informations and secret of API
type App struct {
	tpl         *template.Template
	providers   map[string]provider.Provider
	secretKey   string
	wsConn      *websocket.Conn
	wsErrCount  uint
	workerCalls sync.Map
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
	app := &App{
		tpl: template.Must(template.New(`iot`).Funcs(template.FuncMap{
			`sha`: tools.Sha1,
		}).ParseGlob(`./templates/*.gohtml`)),
		providers:   providers,
		secretKey:   *config[`secretKey`],
		workerCalls: sync.Map{},
	}

	for _, provider := range providers {
		provider.SetHub(app)
	}

	return app
}

// Flags add flags for given prefix
func Flags(prefix string) map[string]*string {
	return map[string]*string{
		`secretKey`: flag.String(tools.ToCamel(fmt.Sprintf(`%sSecretKey`, prefix)), ``, `[iot] Secret Key between worker and API`),
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
		log.Printf(message.Content)
	}

	for name, provider := range a.providers {
		response[name] = provider.GetData()
	}

	w.Header().Set(`content-language`, `fr`)
	if err := templates.WriteHTMLTemplate(a.tpl.Lookup(`iot`), w, response, status); err != nil {
		httperror.InternalServerError(w, err)
	}
}

// Handler create Handler with given App context
func (a *App) Handler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		a.RenderDashboard(w, r, http.StatusOK, nil)
	})
}
