package nec

import (
	"net"
	"strconv"
	"strings"

	event "bitbucket.org/carrierlabs/dev-util-event"
	"github.com/sirupsen/logrus"
)

var log *logrus.Entry

// Display is a representation of a NEC screen
type Display struct {
	// Channels
	eventFeedback chan event.Event
	eventControl  chan event.Event
	// Communications
	conn net.Conn
	ip   struct {
		connected bool
		host      string
		port      int
	}
	id int
	// MetaData
	PartNumber  string
	ModelName   string
	ModelDesc   string
	FirmwareVer string
	// Status
	Input     int
	VideoMute bool
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
	}
	d.ip.port = 7142 // Default Port

	// Configure logger
	if config.Logger == nil {
		log = logrus.New().WithField("package", "dev-display-nec")
	} else {
		log = config.Logger.WithFields(logrus.Fields{"package": "dev-display-nec"})
	}

	// Store ID
	d.id = config.ID
	if d.id == 0 {
		d.id = 1 // Default ID
	}

	// Store Host & Port
	for i, x := range strings.Split(config.Host, `:`) {
		switch i {
		case 0:
			d.ip.host = x
		case 1:
			d.ip.port, _ = strconv.Atoi(x)
		}
	}

	go d.commsLoop()

	return d
}
