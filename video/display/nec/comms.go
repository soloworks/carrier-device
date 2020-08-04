package nec

import (
	"bytes"
	"encoding/hex"
	"fmt"

	event "bitbucket.org/carrierlabs/dev-util-event"
)

// send query to the display
func (d *Display) getCommand(cmd []byte) {
	d.send(0x41, cmd)
}

// send command to the display
func (d *Display) setCommand(cmd []byte, val []byte) {

	// Build the message
	msg := new(bytes.Buffer)  // Create Buffer
	msg.Write(cmd)            // Add Command
	msg.Write(val)            // Add Value
	d.send(0x41, msg.Bytes()) // Send It
}

// Phsyical send function
func (d *Display) send(t byte, m []byte) {

	// Build the message
	msg := new(bytes.Buffer) // Create Buffer
	msg.WriteByte(0x02)      // Add STX
	msg.Write(m)             // Add Message
	msg.WriteByte(0x03)      // Add ETX

	// Build the Packet
	pkt := new(bytes.Buffer)                               // Create Buffer
	pkt.WriteByte(0x01)                                    // Add SOH
	pkt.WriteByte(0x30)                                    // Add Reserved
	pkt.WriteByte(0x40 + byte(d.id))                       // Add Display ID
	pkt.WriteByte(0x30)                                    // Add Message Sender is Controller
	pkt.WriteByte(t)                                       // Add Message type
	pkt.WriteString(fmt.Sprintf("%02X", len(msg.Bytes()))) // Add Message Length (2 char Hex as Ascii)
	pkt.Write(msg.Bytes())                                 // Add Message
	chk := 0x00
	for i, x := range pkt.Bytes() {
		if i > 0 {
			chk = chk ^ int(x)
		}
	}
	pkt.WriteByte(byte(chk))
	pkt.WriteByte(0x0D)

	log.Print("Tx::", hex.Dump(pkt.Bytes()))

	d.conn.Write(pkt.Bytes())

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
