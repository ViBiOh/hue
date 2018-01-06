package provider

import (
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

// ErrorPrefix for sending back error
var ErrorPrefix = []byte(`error `)

// Message rendered to user
type Message struct {
	Level   string
	Content string
}

// Provider for IoT
type Provider interface {
	SetHub(Hub)
	GetWorkerPrefix() []byte
	GetData() interface{}
	WorkerHandler([]byte)
}

// Hub for rendering UI
type Hub interface {
	SendToWorker([]byte) bool
	RenderDashboard(http.ResponseWriter, *http.Request, int, *Message)
}

// WriteTextMessage writes content as text message on websocket
func WriteTextMessage(ws *websocket.Conn, content []byte) bool {
	if err := ws.WriteMessage(websocket.TextMessage, content); err != nil {
		log.Printf(`Error while sending text message %s: %v`, content, err)
		return false
	}
	return true
}

// WriteErrorMessage writes error message on websocket
func WriteErrorMessage(ws *websocket.Conn, errPayload error) bool {
	if err := ws.WriteMessage(websocket.TextMessage, append(ErrorPrefix, []byte(errPayload.Error())...)); err != nil {
		log.Printf(`Error while sending error message %v: %v`, errPayload, err)
		return false
	}
	return true
}
