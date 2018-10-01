package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/ViBiOh/httputils/pkg/opentracing"
	"github.com/ViBiOh/httputils/pkg/rollbar"
	"github.com/ViBiOh/httputils/pkg/tools"
	"github.com/gorilla/websocket"
)

const (
	// WorkerErrorType for sending back error
	WorkerErrorType = `error`
)

// Message rendered to user
type Message struct {
	Level   string
	Content string
}

// WorkerMessage describe how message are exchanged accross worker
type WorkerMessage struct {
	ID      string
	Source  string
	Type    string
	Tracing map[string]string
	Payload interface{}
}

// Provider for IoT
type Provider interface {
	SetHub(Hub)
	GetWorkerSource() string
	GetData(ctx context.Context) interface{}
	WorkerHandler(*WorkerMessage) error
}

// Hub for rendering UI
type Hub interface {
	SendToWorker(context.Context, *WorkerMessage, bool) *WorkerMessage
	RenderDashboard(http.ResponseWriter, *http.Request, int, *Message)
}

// WriteMessage writes content as text message on websocket
func WriteMessage(ctx context.Context, ws *websocket.Conn, message *WorkerMessage) error {
	if ws == nil {
		return fmt.Errorf(`no websocket connection provided for sending: %+v`, message)
	}

	message.Tracing = make(map[string]string)
	if err := opentracing.InjectSpanToMap(ctx, message.Tracing); err != nil {
		rollbar.LogError(`%v`, err)
	}

	messagePayload, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf(`error while marshalling message to worker: %v`, err)
	}

	if err := ws.WriteMessage(websocket.TextMessage, messagePayload); err != nil {
		return fmt.Errorf(`error while sending message to worker: %v`, err)
	}

	return nil
}

// WriteErrorMessage writes error message on websocket
func WriteErrorMessage(ws *websocket.Conn, source string, errPayload error) error {
	message := &WorkerMessage{
		ID:      tools.Sha1(errPayload),
		Source:  source,
		Type:    WorkerErrorType,
		Payload: errPayload,
	}

	messagePayload, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf(`error while marshalling error message: %v`, err)
	}

	if err := ws.WriteMessage(websocket.TextMessage, messagePayload); err != nil {
		return fmt.Errorf(`error while sending error message: %v`, err)
	}

	return nil
}
