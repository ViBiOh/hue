package provider

import "net/http"

// ErrorPrefix for sending back error
var ErrorPrefix = []byte(`error `)

// Message rendered to user
type Message struct {
	Level   string
	Content string
}

// Provider for IoT
type Provider interface {
	SetRenderer(Renderer)
	GetData() interface{}
}

// Renderer for rendering UI
type Renderer interface {
	RenderDashboard(http.ResponseWriter, *http.Request, int, *Message)
}
