package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

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
	SendToWorker(*WorkerMessage, bool) *WorkerMessage
	RenderDashboard(http.ResponseWriter, *http.Request, int, *Message)
}

// WriteMessage writes content as text message on websocket
func WriteMessage(ws *websocket.Conn, message *WorkerMessage) error {
	if ws == nil {
		return fmt.Errorf(`No websocket connection provided for sending: %+v`, message)
	}

	messagePayload, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf(`Error while marshalling message to worker: %v`, err)
	}

	if err := ws.WriteMessage(websocket.TextMessage, messagePayload); err != nil {
		return fmt.Errorf(`Error while sending message to worker: %v`, err)
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
		return fmt.Errorf(`Error while marshalling error message: %v`, err)
	}

	if err := ws.WriteMessage(websocket.TextMessage, messagePayload); err != nil {
		return fmt.Errorf(`Error while sending error message: %v`, err)
	}

	return nil
}
