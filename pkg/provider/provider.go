package provider

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/ViBiOh/httputils/pkg/opentracing"
	"github.com/ViBiOh/httputils/pkg/rollbar"
	"github.com/ViBiOh/httputils/pkg/tools"
	"github.com/gorilla/websocket"
)

const (
	// WorkerErrorAction for sending back error
	WorkerErrorAction = `error`
)

var (
	// ErrWorkerUnknownAction is default error when a worker provider doesn't know how to handle action
	ErrWorkerUnknownAction = errors.New(`unknown action`)
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
	Action  string
	Tracing map[string]string
	Payload string
}

// Hub that renders UI to end user
type Hub interface {
	SendToWorker(ctx context.Context, root *WorkerMessage, source, action string, payload string, waitOutput bool) *WorkerMessage
	RenderDashboard(http.ResponseWriter, *http.Request, int, *Message)
}

// Provider of data for UI
type Provider interface {
	GetData() interface{}
}

// HubUser is a component that need to interact directly with the Hub
type HubUser interface {
	SetHub(Hub)
}

// WorkerProvider is a provider that need to interact with the remote Worker
type WorkerProvider interface {
	GetWorkerSource() string
	WorkerHandler(*WorkerMessage) error
}

// Worker is a remote worker in another network, connected with websocket to hub
type Worker interface {
	GetSource() string
	Handle(context.Context, *WorkerMessage) (*WorkerMessage, error)
	Ping(context.Context) ([]*WorkerMessage, error)
}

// NewWorkerMessage instantiates a worker message
func NewWorkerMessage(root *WorkerMessage, source, action string, payload string) *WorkerMessage {
	var id string
	if root == nil || strings.TrimSpace(root.ID) == `` {
		id = tools.Sha1(payload)
	} else {
		id = root.ID
	}

	return &WorkerMessage{
		ID:      id,
		Source:  source,
		Action:  action,
		Payload: payload,
	}
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
	message := NewWorkerMessage(nil, source, WorkerErrorAction, errPayload.Error())

	messagePayload, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf(`error while marshalling error message: %v`, err)
	}

	if err := ws.WriteMessage(websocket.TextMessage, messagePayload); err != nil {
		return fmt.Errorf(`error while sending error message: %v`, err)
	}

	return nil
}
