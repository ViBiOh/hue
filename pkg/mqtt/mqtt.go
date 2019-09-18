package mqtt

import (
	"crypto/tls"
	"flag"
	"fmt"

	"github.com/ViBiOh/httputils/v2/pkg/errors"
	"github.com/ViBiOh/httputils/v2/pkg/logger"
	"github.com/ViBiOh/httputils/v2/pkg/tools"
	"github.com/yosssi/gmq/mqtt"
	"github.com/yosssi/gmq/mqtt/client"
)

// Config of package
type Config struct {
	server   *string
	port     *int
	useTLS   *bool
	user     *string
	pass     *string
	clientID *string
}

// App of package
type App struct {
	client *client.Client
}

// Flags adds flags for configuring package
func Flags(fs *flag.FlagSet, prefix string) Config {
	return Config{
		server:   tools.NewFlag(prefix, "mqtt").Name("Server").Default("").Label("Server name").ToString(fs),
		port:     tools.NewFlag(prefix, "mqtt").Name("Port").Default(80).Label("Port").ToInt(fs),
		useTLS:   tools.NewFlag(prefix, "mqtt").Name("UseTLS").Default(true).Label("Use TLS").ToBool(fs),
		user:     tools.NewFlag(prefix, "mqtt").Name("User").Default("").Label("Username").ToString(fs),
		pass:     tools.NewFlag(prefix, "mqtt").Name("Pass").Default("").Label("Password").ToString(fs),
		clientID: tools.NewFlag(prefix, "mqtt").Name("ClientID").Default("iot").Label("Client ID").ToString(fs),
	}
}

// New creates new App from Config
func New(config Config) (*App, error) {
	if *config.server == "" {
		logger.Warn("no server provided")
		return &App{}, nil
	}

	if *config.clientID == "" {
		logger.Warn("no clientID provided")
		return &App{}, nil
	}

	app, err := Connect(*config.server, *config.user, *config.pass, *config.clientID, *config.port, *config.useTLS)
	if err != nil {
		return nil, err
	}

	return app, nil
}

// Connect to MQTT
func Connect(server, user, pass, clientID string, port int, useTLS bool) (*App, error) {
	var tlsConfig *tls.Config
	if useTLS {
		tlsConfig = &tls.Config{}
	}

	connect := func(mqttClient *client.Client) error {
		return errors.WithStack(mqttClient.Connect(&client.ConnectOptions{
			Network:   "tcp",
			Address:   fmt.Sprintf("%s:%d", server, port),
			TLSConfig: tlsConfig,
			UserName:  []byte(user),
			Password:  []byte(pass),
			ClientID:  []byte(clientID),
		}))
	}

	var mqttClient *client.Client

	handleError := func(err error) {
		logger.Error("error with %s as %s: %#v", server, clientID, err)
		if err == client.ErrNotYetConnected {
			if err := connect(mqttClient); err != nil {
				logger.Error("error while attempting to reconnect: %#v", err)
			}
		}
	}

	mqttClient = client.New(&client.Options{
		ErrorHandler: handleError,
	})

	err := connect(mqttClient)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return &App{mqttClient}, nil
}

// Enabled determines if MQTT is enabled or not
func (a App) Enabled() bool {
	return a.client != nil
}

// Publish to a topic
func (a App) Publish(topic string, message []byte) error {
	if !a.Enabled() {
		return errors.New("client not configured")
	}

	err := a.client.Publish(&client.PublishOptions{
		QoS:       mqtt.QoS0,
		Retain:    true,
		TopicName: []byte(topic),
		Message:   message,
	})

	return errors.WithStack(err)
}

// Subscribe to a topic
func (a App) Subscribe(topic string, handler func([]byte)) error {
	if !a.Enabled() {
		return errors.New("client not configured")
	}

	err := a.client.Subscribe(&client.SubscribeOptions{
		SubReqs: []*client.SubReq{
			{
				TopicFilter: []byte(topic),
				QoS:         mqtt.QoS0,
				Handler: func(_, message []byte) {
					handler(message)
				},
			},
		},
	})

	return errors.WithStack(err)
}

// Unsubscribe from a topic
func (a App) Unsubscribe(topic string) error {
	if !a.Enabled() {
		return errors.New("client not configured")
	}

	err := a.client.Unsubscribe(&client.UnsubscribeOptions{
		TopicFilters: [][]byte{
			[]byte(topic),
		},
	})

	return errors.WithStack(err)
}

// End disconnect and release ressource properly
func (a App) End() {
	if !a.Enabled() {
		return
	}

	if err := a.client.Disconnect(); err != nil {
		logger.Error("%#v", errors.WithStack(err))
	}

	a.client.Terminate()
}
