package nec

import (
	"bufio"
	"encoding/hex"
	"net"
	"strconv"
	"strings"
	"time"

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

// SetSource sets source s on destination d
// d is ignored in this device
func (d *Display) SetSource(s int, dest int) {
	d.conn.Write([]byte(strconv.Itoa(s) + "!"))
}

// GetSource returns current input number
func (d *Display) GetSource() int {
	return d.Input
}

// SetPower is unsued on this device
func (d *Display) SetPower(p bool) {
	switch p {
	case true:
		d.setCommand([]byte("C203D6"), []byte("0001"))
	case false:
		d.setCommand([]byte("C203D6"), []byte("0004"))
	}
}

// GetPower always returns true on this device
func (d *Display) GetPower() bool { return true }

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
	if config.ID == 0 {
		d.id = 1 // Default ID
	} else {
		d.id = config.ID
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

	return nil
}

func (d *Display) commsLoop() {
	for {
		var err error
		d.conn, err = net.Dial("tcp", d.ip.host+`:`+strconv.Itoa(+d.ip.port))
		if err != nil {
			log.Errorf("Failed to connect: %v :Waiting to retry", err.Error())
			time.Sleep(time.Millisecond * time.Duration(2000))
		} else {
			log.Info("Connected")
			// Create new Reader
			r := bufio.NewReader(d.conn)

			// Init Device
			d.getCommand([]byte("01D6")) // Query Power State

			for {
				message, err := r.ReadBytes('\x0D')
				if err != nil {
					log.Println("RxErr::", err)
					break
				}
				log.Print("Rx::", hex.Dump(message))

				// Process Feedback

			}
		}
	}
}
