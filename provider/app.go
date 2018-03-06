package provider

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/ViBiOh/iot/utils"
	"github.com/gorilla/websocket"
)

// ErrorPrefix for sending back error
const (
	ErrorPrefix = `error`
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
	GetData() interface{}
	WorkerHandler(*WorkerMessage) error
}

// Hub for rendering UI
type Hub interface {
	SendToWorker(*WorkerMessage) bool
	RenderDashboard(http.ResponseWriter, *http.Request, int, *Message)
}

// WriteMessage writes content as text message on websocket
func WriteMessage(ws *websocket.Conn, message *WorkerMessage) bool {
	if ws == nil {
		log.Printf(`No websocket connection provided for sending: %+v`, message)
		return false
	}

	messagePayload, err := json.Marshal(message)
	if err != nil {
		log.Printf(`Error while marshalling message to worker: %v`, err)
		return false
	}

	if err := ws.WriteMessage(websocket.TextMessage, messagePayload); err != nil {
		log.Printf(`Error while sending message to worker: %v`, err)
		return false
	}

	return true
}

// WriteErrorMessage writes error message on websocket
func WriteErrorMessage(ws *websocket.Conn, errPayload error) bool {
	message := &WorkerMessage{
		ID:      utils.ShaFingerprint(errPayload),
		Source:  ErrorPrefix,
		Type:    ErrorPrefix,
		Payload: errPayload,
	}

	messagePayload, err := json.Marshal(message)
	if err != nil {
		log.Printf(`Error while marshalling error message: %v`, err)
		return false
	}

	if err := ws.WriteMessage(websocket.TextMessage, messagePayload); err != nil {
		log.Printf(`Error while sending error message: %v`, err)
		return false
	}

	return true
}
