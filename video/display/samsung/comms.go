package samsung

import (
	"bufio"
	"encoding/hex"
	"net"
	"strconv"
	"time"

	event "bitbucket.org/carrierlabs/dev-util-event"
)

type comms struct {
	// Communications
	conn      net.Conn
	connected bool
	host      string // Network Host
	port      int    // Network Control Port
	id        int    // Display ID
	tx        struct {
		cmd [][]byte // Transmit Queue for Commands
		qry [][]byte // Transmit Queue for Queries
	}
}

// SendQuery packs and queues a query
func (d *Display) SendQuery(cmd byte) {

}

// Send packs and queues a query
func (d *Display) Send(cmd byte, data ...byte) {

	if len(data) == 0 {
		d.comms.tx.qry = append(d.comms.tx.qry, packMessage(d.comms.id, cmd))
	} else {
		d.comms.tx.cmd = append(d.comms.tx.cmd, packMessage(d.comms.id, cmd, data...))
	}

}

// EventFeedback returns a read-only channel which emits events as they occur on the
func (d *Display) EventFeedback() <-chan event.Event {
	return d.eventFeedback
}

// EventControl returns a write-only channel for sending control events to the device
func (d *Display) EventControl() chan<- event.Event {
	return d.eventFeedback
}
func (d *Display) commsLoop() {
	for {
		var err error
		d.comms.conn, err = net.Dial("tcp", d.comms.host+`:`+strconv.Itoa(+d.comms.port))
		if err != nil {
			log.Println("Failed to connect:", err.Error())
			log.Println("Trying reset the connectiod...")
			time.Sleep(time.Millisecond * time.Duration(2000))
		} else {
			log.Println("Connected")
			// Create new Reader
			r := bufio.NewReader(d.comms.conn)

			d.Send(0x11) // Query Power State

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
