package iot

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

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
	iotSource        = `iot`
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
		}).ParseGlob(`./web/*.gohtml`)),
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
		`secretKey`: flag.String(tools.ToCamel(fmt.Sprintf(`%s%s`, prefix, `SecretKey`)), ``, `[iot] Secret Key between worker and API`),
	}
}

func (a *App) checkWorker(ws *websocket.Conn) bool {
	messageType, p, err := ws.ReadMessage()

	if err != nil {
		if err := provider.WriteErrorMessage(ws, iotSource, fmt.Errorf(`Error while reading first message: %v`, err)); err != nil {
			log.Print(err)
		}
		return false
	}

	if messageType != websocket.TextMessage {
		if err := provider.WriteErrorMessage(ws, iotSource, errors.New(`First message should be a Text Message`)); err != nil {
			log.Print(err)
		}
		return false
	}

	if string(p) != a.secretKey {
		if err := provider.WriteErrorMessage(ws, iotSource, errors.New(`First message should be the Secret Key`)); err != nil {
			log.Print(err)
		}
		return false
	}

	return true
}

// SendToWorker sends payload to worker
func (a *App) SendToWorker(message *provider.WorkerMessage, waitOutput bool) *provider.WorkerMessage {
	var outputChan chan *provider.WorkerMessage

	if waitOutput {
		outputChan = make(chan *provider.WorkerMessage)
		a.workerCalls.Store(message.ID, outputChan)

		defer a.workerCalls.Delete(message.ID)
	}

	if err := provider.WriteMessage(a.wsConn, message); err != nil {
		return &provider.WorkerMessage{
			Source:  message.Source,
			Type:    provider.WorkerErrorType,
			Payload: err,
		}
	}

	if waitOutput {
		select {
		case output := <-outputChan:
			return output
		case <-time.After(10 * time.Second):
			return nil
		}
	}

	return nil
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
					log.Printf(`[iot] Error while closing connection: %v`, err)
				}
			}()
		}
		if err != nil {
			log.Printf(`[iot] Error while upgrading connection: %v`, err)
			return
		}

		if !a.checkWorker(ws) {
			return
		}

		log.Printf(`Worker connection from %s`, request.GetIP(r))
		if a.wsConn != nil {
			if err := a.wsConn.Close(); err != nil {
				log.Printf(`[iot] Error while closing connection: %v`, err)
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
				log.Printf(`[iot] Error while reading from websocket: %v`, err)
				return
			}

			if messageType == websocket.TextMessage {
				var workerMessage provider.WorkerMessage
				if err := json.Unmarshal(p, &workerMessage); err != nil {
					log.Printf(`[iot] Error while unmarshalling worker message: %v`, err)
					a.wsErrCount++
					break
				}

				if outputChan, ok := a.workerCalls.Load(workerMessage.ID); ok {
					outputChan.(chan *provider.WorkerMessage) <- &workerMessage
				}

				if workerMessage.Type == provider.WorkerErrorType {
					log.Printf(`[iot] [%s] %v`, workerMessage.Source, workerMessage.Payload)
					break
				}

				for name, value := range a.providers {
					if strings.HasPrefix(workerMessage.Source, value.GetWorkerSource()) {
						if err := value.WorkerHandler(&workerMessage); err != nil {
							log.Printf(`[iot] [%s] %v`, name, err)
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
