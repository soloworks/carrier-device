package samsung

import (
	conn "bitbucket.org/carrierlabs/dev-util-conn"
	event "bitbucket.org/carrierlabs/dev-util-event"
	"github.com/sirupsen/logrus"
)

var log *logrus.Entry

type metaData struct {
	SerialNumber string
	ModelName    string
}
type state struct {
	Input int
}

// Display is a representation of a NEC screen
type Display struct {
	// Channels
	eventFeedback chan event.Event
	eventControl  chan event.Event
	conn          *conn.Conn
	metaData      metaData // MetaData
	state         state    // Status
	id            int
}

// Config allows fields to be set to configure a new instance
type Config struct {
	Host   string
	ID     int
	Logger *logrus.Logger
}

// New returns a new Device instance
func New(config Config) *Display {
	d := &Display{
		eventFeedback: make(chan event.Event, 50),
		eventControl:  make(chan event.Event, 50),
	}

	// Store ID
	if d.id == 0 {
		d.id = 1 // Default ID
	}

	// Configure logger
	lf := logrus.Fields{
		"package": "dev-display-samsung",
		"host":    config.Host,
		"id":      d.id,
	}
	if config.Logger == nil {
		log = logrus.New().WithFields(lf)
	} else {
		log = config.Logger.WithFields(lf)
	}

	// Spin up a new device core
	var err error
	d.conn, err = conn.New(conn.Config{
		Host:              config.Host,
		Port:              1515,
		Encoder:           d.encoder,
		Decoder:           d.decoder,
		ConnectionTimeout: 20,
		ResponseTimeout:   5,
	}) // Generate a new device connection
	if err != nil {
		log.Error(err)
	}

	return d
}

// EventFeedback returns a read-only channel which emits events as they occur on the
// base server
func (d *Display) EventFeedback() <-chan event.Event {
	return d.eventFeedback
}

// EventControl returns a write-only channel for sending control events to the device
func (d *Display) EventControl() chan<- event.Event {
	return d.eventFeedback
}
