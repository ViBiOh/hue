package iot

import (
	"bytes"
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"flag"
	"fmt"
	"html/template"
	"log"
	"net/http"

	"github.com/ViBiOh/httputils/httperror"
	"github.com/ViBiOh/httputils/request"
	"github.com/ViBiOh/httputils/templates"
	"github.com/ViBiOh/httputils/tools"
	"github.com/ViBiOh/iot/provider"
	"github.com/gorilla/websocket"
)

const (
	maxAllowedErrors = 5
	hoursInDay       = 24
	minutesInHours   = 60
)

var (
	hours   []string
	minutes []string
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// App stores informations and secret of API
type App struct {
	tpl        *template.Template
	providers  map[string]provider.Provider
	secretKey  string
	wsConn     *websocket.Conn
	wsErrCount uint
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
			`sha`: func(content interface{}) string {
				hasher := sha1.New()
				if _, err := hasher.Write([]byte(fmt.Sprintf(`%v`, content))); err != nil {
					log.Printf(`Error while generating hash for %s: %v`, content, err)
				}

				return hex.EncodeToString(hasher.Sum(nil))
			},
		}).ParseGlob(`./web/*.gohtml`)),
		providers: providers,
		secretKey: *config[`secretKey`],
	}

	for _, provider := range providers {
		provider.SetHub(app)
	}

	return app
}

// Flags add flags for given prefix
func Flags(prefix string) map[string]*string {
	return map[string]*string{
		`secretKey`: flag.String(tools.ToCamel(prefix+`SecretKey`), ``, `[iot] Secret Key between worker and API`),
	}
}

func (a *App) checkWorker(ws *websocket.Conn) bool {
	messageType, p, err := ws.ReadMessage()

	if err != nil {
		provider.WriteErrorMessage(ws, fmt.Errorf(`Error while reading first message: %v`, err))
		return false
	}

	if messageType != websocket.TextMessage {
		provider.WriteErrorMessage(ws, errors.New(`First message should be a Text Message`))
		return false
	}

	if string(p) != a.secretKey {
		provider.WriteErrorMessage(ws, errors.New(`First message should be the Secret Key`))
		return false
	}

	return true
}

// SendToWorker sends payload to worker
func (a *App) SendToWorker(payload []byte) bool {
	return provider.WriteTextMessage(a.wsConn, payload)
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

// WebsocketHandler create Websockethandler
func (a *App) WebsocketHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ws, err := upgrader.Upgrade(w, r, nil)
		if ws != nil {
			defer func() {
				if a.wsConn == ws {
					a.wsConn = nil
				}

				if err := ws.Close(); err != nil {
					log.Printf(`Error while closing connection: %v`, err)
				}
			}()
		}
		if err != nil {
			log.Printf(`Error while upgrading connection: %v`, err)
			return
		}

		if !a.checkWorker(ws) {
			return
		}

		log.Printf(`Worker connection from %s`, request.GetIP(r))
		if a.wsConn != nil {
			if err := a.wsConn.Close(); err != nil {
				log.Printf(`Error while closing connection: %v`, err)
			}

		}
		a.wsConn = ws

		a.wsErrCount = 0

		for {
			messageType, p, err := ws.ReadMessage()
			if messageType == websocket.CloseMessage {
				return
			}

			if err != nil {
				log.Printf(`Error while reading from websocket: %v`, err)
				return
			}

			if messageType == websocket.TextMessage {
				if bytes.HasPrefix(p, provider.ErrorPrefix) {
					log.Printf(`Error received from worker: %s`, bytes.TrimPrefix(p, provider.ErrorPrefix))
					a.wsErrCount++
					break
				}

				for name, value := range a.providers {
					if bytes.HasPrefix(p, value.GetWorkerPrefix()) {
						if err := value.WorkerHandler(bytes.TrimPrefix(p, value.GetWorkerPrefix())); err != nil {
							log.Printf(`[%s] %v`, name, err)
						}
						a.wsErrCount = 0
						break
					}
				}
			}
		}
	})
}

// Handler create Handler with given App context
func (a *App) Handler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		a.RenderDashboard(w, r, http.StatusOK, nil)
	})
}
