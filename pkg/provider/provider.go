package provider

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/ViBiOh/httputils/v2/pkg/errors"
	"github.com/ViBiOh/httputils/v2/pkg/logger"
	"github.com/ViBiOh/httputils/v2/pkg/opentracing"
	"github.com/ViBiOh/httputils/v2/pkg/tools"
	"github.com/ViBiOh/iot/pkg/mqtt"
)

const (
	// WorkerErrorAction for sending back error
	WorkerErrorAction = "error"
)

var (
	// ErrWorkerUnknownAction is default error when a worker provider doesn't know how to handle action
	ErrWorkerUnknownAction = errors.New("unknown action")
)

// Message rendered to user
type Message struct {
	Level   string
	Content string
}

// WorkerMessage describe how message are exchanged across worker
type WorkerMessage struct {
	ID         string
	ResponseTo string
	Source     string
	Action     string
	Tracing    map[string]string
	Payload    string
}

// Hub that renders UI to end user
type Hub interface {
	SendToWorker(ctx context.Context, root *WorkerMessage, source, action string, payload string, waitOutput bool) *WorkerMessage
	RenderDashboard(http.ResponseWriter, *http.Request, int, *Message)
}

// Provider of data for UI
type Provider interface {
	EnablePrometheus()
	GetData() interface{}
}

// HubUser is a component that need to interact directly with the Hub
type HubUser interface {
	SetHub(Hub)
}

// WorkerProvider is a provider that interact with the remote Worker
type WorkerProvider interface {
	GetWorkerSource() string
	WorkerHandler(*WorkerMessage) error
}

// Worker is a remote worker in another network, connected to hub
type Worker interface {
	GetSource() string
	Enabled() bool
}

// Pinger is a Worker that respond to frequent Ping
type Pinger interface {
	Ping(context.Context) ([]*WorkerMessage, error)
}

// WorkerHandler is a Worker that can handle request from outside
type WorkerHandler interface {
	Handle(context.Context, *WorkerMessage) (*WorkerMessage, error)
}

// Starter is a component that need to be started
type Starter interface {
	Start()
	Enabled() bool
}

// NewWorkerMessage instantiates a worker message
func NewWorkerMessage(root *WorkerMessage, source, action string, payload string) *WorkerMessage {
	var id string
	if root == nil || strings.TrimSpace(root.ID) == "" {
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

// WriteMessage writes content as text message
func WriteMessage(ctx context.Context, client *mqtt.App, topic string, message *WorkerMessage) error {
	if client == nil {
		return errors.New("no connection provided for sending: %#v", message)
	}

	message.Tracing = make(map[string]string)
	if err := opentracing.InjectSpanToMap(ctx, message.Tracing); err != nil {
		logger.Error("%#v", errors.WithStack(err))
	}

	messagePayload, err := json.Marshal(message)
	if err != nil {
		return errors.WithStack(err)
	}

	if err := client.Publish(topic, messagePayload); err != nil {
		return errors.WithStack(err)
	}

	return nil
}
