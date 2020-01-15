package nec

import (
	"bufio"
	"bytes"
	"encoding/hex"
	"fmt"
	"log"
	"net"
	"strconv"
	"strings"
	"time"
)

type nec struct {
	conn      net.Conn
	ipHost    string
	ipPort    int
	id        int
	connected bool
	// MetaData
	PartNumber  string
	ModelName   string
	ModelDesc   string
	FirmwareVer string
	// Status
	Input        int
	VideoMute    bool
	SignalInput  []bool
	SignalOutput bool
}

// send pushed command to the switcher
func (n *nec) getCommand(cmd []byte) {
	n.send(0x41, cmd)
}

// send pushed command to the switcher
func (n *nec) setCommand(cmd []byte, val []byte) {

	// Build the message
	msg := new(bytes.Buffer)  // Create Buffer
	msg.Write(cmd)            // Add Command
	msg.Write(val)            // Add Value
	n.send(0x41, msg.Bytes()) // Send It
}

func (n *nec) send(t byte, m []byte) {

	// Build the message
	msg := new(bytes.Buffer) // Create Buffer
	msg.WriteByte(0x02)      // Add STX
	msg.Write(m)             // Add Message
	msg.WriteByte(0x03)      // Add ETX

	// Build the Packet
	pkt := new(bytes.Buffer)                               // Create Buffer
	pkt.WriteByte(0x01)                                    // Add SOH
	pkt.WriteByte(0x30)                                    // Add Reserved
	pkt.WriteByte(0x40 + byte(n.id))                       // Add Display ID
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

	n.conn.Write(pkt.Bytes())

}

// SetSource sets source s on destination d
// d is ignored in this device
func (n *nec) SetSource(s int, d int) {
	n.conn.Write([]byte(strconv.Itoa(s) + "!"))
}

// GetSource returns current input number
func (n *nec) GetSource() int {
	return n.Input
}

// SetPower is unsued on this device
func (n *nec) SetPower(p bool) {
	switch p {
	case true:
		n.setCommand([]byte("C203D6"), []byte("0001"))
	case false:
		n.setCommand([]byte("C203D6"), []byte("0004"))
	}
}

// Getpower always returns true on this device
func (n *nec) GetPower() bool { return true }

// New creates a new instance, and initialises communication
func (n *nec) Init(ip string) error {

	n.id = 1        // Default ID
	n.ipPort = 7142 // Default Port

	// Store Host & Port
	for i, x := range strings.Split(ip, `:`) {
		switch i {
		case 0:
			n.ipHost = x
		case 1:
			n.ipPort, _ = strconv.Atoi(x)
		}
	}

	go func() {
		for {
			var err error
			n.conn, err = net.Dial("tcp", n.ipHost+`:`+strconv.Itoa(+n.ipPort))
			if err != nil {
				log.Println("Failed to connect:", err.Error())
				log.Println("Trying reset the connection...")
				time.Sleep(time.Millisecond * time.Duration(2000))
			} else {
				log.Println("Connected")
				// Create new Reader
				r := bufio.NewReader(n.conn)

				// Init Device
				n.getCommand([]byte("01D6")) // Query Power State

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
	}()
	return nil
}
