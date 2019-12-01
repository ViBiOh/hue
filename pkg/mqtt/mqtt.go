package mqtt

import (
	"crypto/tls"
	"errors"
	"flag"
	"fmt"

	"github.com/ViBiOh/httputils/v3/pkg/flags"
	"github.com/ViBiOh/httputils/v3/pkg/logger"
	"github.com/yosssi/gmq/mqtt"
	"github.com/yosssi/gmq/mqtt/client"
)

// App of package
type App interface {
	Enabled() bool
	Publish(string, []byte) error
	Subscribe(string, func([]byte)) error
	Unsubscribe(string) error
}

// Config of packag
type Config struct {
	server   *string
	port     *int
	useTLS   *bool
	user     *string
	pass     *string
	clientID *string
}

type app struct {
	client *client.Client
}

// Flags adds flags for configuring package
func Flags(fs *flag.FlagSet, prefix string) Config {
	return Config{
		server:   flags.New(prefix, "mqtt").Name("Server").Default("").Label("Server name").ToString(fs),
		port:     flags.New(prefix, "mqtt").Name("Port").Default(80).Label("Port").ToInt(fs),
		useTLS:   flags.New(prefix, "mqtt").Name("UseTLS").Default(true).Label("Use TLS").ToBool(fs),
		user:     flags.New(prefix, "mqtt").Name("User").Default("").Label("Username").ToString(fs),
		pass:     flags.New(prefix, "mqtt").Name("Pass").Default("").Label("Password").ToString(fs),
		clientID: flags.New(prefix, "mqtt").Name("ClientID").Default("iot").Label("Client ID").ToString(fs),
	}
}

// New creates new App from Config
func New(config Config) (App, error) {
	if *config.server == "" {
		logger.Warn("no server provided")
		return &app{}, nil
	}

	if *config.clientID == "" {
		logger.Warn("no clientID provided")
		return &app{}, nil
	}

	return Connect(*config.server, *config.user, *config.pass, *config.clientID, *config.port, *config.useTLS)
}

// Connect to MQTT
func Connect(server, user, pass, clientID string, port int, useTLS bool) (App, error) {
	var tlsConfig *tls.Config
	if useTLS {
		tlsConfig = &tls.Config{}
	}

	connect := func(mqttClient *client.Client) error {
		return mqttClient.Connect(&client.ConnectOptions{
			Network:   "tcp",
			Address:   fmt.Sprintf("%s:%d", server, port),
			TLSConfig: tlsConfig,
			UserName:  []byte(user),
			Password:  []byte(pass),
			ClientID:  []byte(clientID),
		})
	}

	var mqttClient *client.Client
	connected := false

	handleError := func(err error) {
		logger.Error("error with %s as %s: %s", server, clientID, err)
		if connected {
			if err := connect(mqttClient); err != nil {
				logger.Error("error while attempting to reconnect: %s", err)
			}
		}
	}

	mqttClient = client.New(&client.Options{
		ErrorHandler: handleError,
	})

	err := connect(mqttClient)
	if err != nil {
		return nil, err
	}

	connected = true
	return &app{mqttClient}, nil
}

// Enabled determines if MQTT is enabled or not
func (a app) Enabled() bool {
	return a.client != nil
}

// Publish to a topic
func (a app) Publish(topic string, message []byte) error {
	if !a.Enabled() {
		return errors.New("client not configured")
	}

	err := a.client.Publish(&client.PublishOptions{
		QoS:       mqtt.QoS0,
		Retain:    true,
		TopicName: []byte(topic),
		Message:   message,
	})

	return err
}

// Subscribe to a topic
func (a app) Subscribe(topic string, handler func([]byte)) error {
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

	return err
}

// Unsubscribe from a topic
func (a app) Unsubscribe(topic string) error {
	if !a.Enabled() {
		return errors.New("client not configured")
	}

	err := a.client.Unsubscribe(&client.UnsubscribeOptions{
		TopicFilters: [][]byte{
			[]byte(topic),
		},
	})

	return err
}

// End disconnect and release ressource properly
func (a app) End() {
	if !a.Enabled() {
		return
	}

	if err := a.client.Disconnect(); err != nil {
		logger.Error("%s", err)
	}

	a.client.Terminate()
}
