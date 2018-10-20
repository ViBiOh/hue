package iot

import (
	"encoding/json"
	"net/http"

	"github.com/ViBiOh/httputils/pkg/errors"
	"github.com/ViBiOh/httputils/pkg/logger"
	"github.com/ViBiOh/httputils/pkg/request"
	"github.com/ViBiOh/iot/pkg/provider"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func (a *App) checkWorker(ws *websocket.Conn) bool {
	messageType, p, err := ws.ReadMessage()

	if err != nil {
		if err := provider.WriteErrorMessage(ws, iotSource, errors.WithStack(err)); err != nil {
			logger.Error(`%+v`, err)
		}
		return false
	}

	if messageType != websocket.TextMessage {
		if err := provider.WriteErrorMessage(ws, iotSource, errors.New(`first message should be a Text Message`)); err != nil {
			logger.Error(`%+v`, err)
		}
		return false
	}

	if string(p) != a.secretKey {
		if err := provider.WriteErrorMessage(ws, iotSource, errors.New(`first message should contains the Secret Key`)); err != nil {
			logger.Error(`%+v`, err)
		}
		return false
	}

	return true
}

func (a *App) handleTextMessage(p []byte) error {
	var workerMessage provider.WorkerMessage
	if err := json.Unmarshal(p, &workerMessage); err != nil {
		a.wsErrCount++
		return errors.WithStack(err)
	}

	if outputChan, ok := a.workerCalls.Load(workerMessage.ID); ok {
		outputChan.(chan *provider.WorkerMessage) <- &workerMessage
	}

	if workerMessage.Action == provider.WorkerErrorAction {
		return errors.New(`%s: %v`, workerMessage.Source, workerMessage.Payload)
	}

	if workerProvider, ok := a.workerProviders[workerMessage.Source]; ok {
		if err := workerProvider.WorkerHandler(&workerMessage); err != nil {
			return err
		}

		a.wsErrCount = 0
		return nil
	}

	return errors.New(`no provider found for message: %+v`, workerMessage)
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
					logger.Error(`%+v`, errors.WithStack(err))
				}
			}()
		}
		if err != nil {
			logger.Error(`%+v`, errors.WithStack(err))
			return
		}

		if !a.checkWorker(ws) {
			return
		}

		logger.Info(`Worker connection from %s`, request.GetIP(r))
		if a.wsConn != nil {
			if err := a.wsConn.Close(); err != nil {
				logger.Error(`%+v`, errors.WithStack(err))
			}

		}

		a.wsConn = ws
		a.wsErrCount = 0

		for {
			messageType, p, err := ws.ReadMessage()
			if messageType == websocket.CloseMessage {
				break
			}

			if err != nil {
				logger.Error(`%+v`, errors.WithStack(err))
				break
			}

			if messageType == websocket.TextMessage {
				if err := a.handleTextMessage(p); err != nil {
					logger.Error(`%+v`, err)
				}
			}
		}
	})
}
