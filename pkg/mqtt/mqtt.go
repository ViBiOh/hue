package mqtt

import (
	"crypto/tls"
	"flag"
	"fmt"

	"github.com/ViBiOh/httputils/pkg/logger"
	"github.com/ViBiOh/httputils/pkg/tools"
	"github.com/pkg/errors"
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
		server:   fs.String(tools.ToCamel(fmt.Sprintf(`%sServer`, prefix)), ``, `[mqtt] Server name`),
		port:     fs.Int(tools.ToCamel(fmt.Sprintf(`%sPort`, prefix)), 80, `[mqtt] Port`),
		useTLS:   fs.Bool(tools.ToCamel(fmt.Sprintf(`%sUseTLS`, prefix)), true, `[mqtt] Use TLS`),
		user:     fs.String(tools.ToCamel(fmt.Sprintf(`%sUser`, prefix)), ``, `[mqtt] Username`),
		pass:     fs.String(tools.ToCamel(fmt.Sprintf(`%sPass`, prefix)), ``, `[mqtt] Password`),
		clientID: fs.String(tools.ToCamel(fmt.Sprintf(`%sClientID`, prefix)), `iot`, `[mqtt] Client ID`),
	}
}

// New creates new App from Config
func New(config Config) (*App, error) {
	if *config.server == `` {
		logger.Warn(`no server provided`)
		return &App{}, nil
	}

	if *config.clientID == `` {
		logger.Warn(`no clientID provided`)
		return &App{}, nil
	}

	mqttClient := client.New(&client.Options{
		ErrorHandler: func(err error) {
			logger.Error(`%+v`, err)
		},
	})

	var tlsConfig *tls.Config
	if *config.useTLS {
		tlsConfig = &tls.Config{}
	}

	if err := mqttClient.Connect(&client.ConnectOptions{
		Network:   `tcp`,
		Address:   fmt.Sprintf(`%s:%d`, *config.server, *config.port),
		TLSConfig: tlsConfig,
		UserName:  []byte(*config.user),
		Password:  []byte(*config.pass),
		ClientID:  []byte(*config.clientID),
	}); err != nil {
		return nil, errors.WithStack(err)
	}

	return &App{
		client: mqttClient,
	}, nil
}

// Publish to a topic
func (a App) Publish(topic string, message []byte) error {
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
	err := a.client.Unsubscribe(&client.UnsubscribeOptions{
		TopicFilters: [][]byte{
			[]byte(topic),
		},
	})

	return errors.WithStack(err)
}

// End disconnect and release ressource properly
func (a App) End() {
	if a.client == nil {
		return
	}

	if err := a.client.Disconnect(); err != nil {
		logger.Error(`%+v`, errors.WithStack(err))
	}

	a.client.Terminate()
}
