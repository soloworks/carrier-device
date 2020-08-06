package samsung

import (
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
	comms         comms
	metaData      metaData // MetaData
	state         state    // Status
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
		eventFeedback: make(chan event.Event),
		eventControl:  make(chan event.Event),
		comms: comms{
			id:   config.ID,
			host: config.Host,
			port: 1515,
		},
	}

	// Configure logger
	lf := logrus.Fields{
		"package": "dev-display-samsung",
		"host":    d.comms.host,
	}
	if config.Logger == nil {
		log = logrus.New().WithFields(lf)
	} else {
		log = config.Logger.WithFields(lf)
	}

	// Store ID
	if d.comms.id == 0 {
		d.comms.id = 1 // Default ID
	}

	go d.commsLoop()

	return d
}
