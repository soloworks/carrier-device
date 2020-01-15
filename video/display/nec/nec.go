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

// Device is a representation of a NEC screen
type Device struct {
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

// send pushed command to the switcher
func (d *Device) getCommand(cmd []byte) {
	d.send(0x41, cmd)
}

// send pushed command to the switcher
func (d *Device) setCommand(cmd []byte, val []byte) {

	// Build the message
	msg := new(bytes.Buffer)  // Create Buffer
	msg.Write(cmd)            // Add Command
	msg.Write(val)            // Add Value
	d.send(0x41, msg.Bytes()) // Send It
}

func (d *Device) send(t byte, m []byte) {

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

// GetHost returns current host
func (d *Device) GetHost() string {
	return d.ip.host
}

// SetSource sets source s on destination d
// d is ignored in this device
func (d *Device) SetSource(s int, dest int) {
	d.conn.Write([]byte(strconv.Itoa(s) + "!"))
}

// GetSource returns current input number
func (d *Device) GetSource() int {
	return d.Input
}

// SetPower is unsued on this device
func (d *Device) SetPower(p bool) {
	switch p {
	case true:
		d.setCommand([]byte("C203D6"), []byte("0001"))
	case false:
		d.setCommand([]byte("C203D6"), []byte("0004"))
	}
}

// GetPower always returns true on this device
func (d *Device) GetPower() bool { return true }

// Init sets defaults, and initialises communication
func (d *Device) Init(ip string) error {

	d.id = 1         // Default ID
	d.ip.port = 7142 // Default Port

	// Store Host & Port
	for i, x := range strings.Split(ip, `:`) {
		switch i {
		case 0:
			d.ip.host = x
		case 1:
			d.ip.port, _ = strconv.Atoi(x)
		}
	}

	go func() {
		for {
			var err error
			d.conn, err = net.Dial("tcp", d.ip.host+`:`+strconv.Itoa(+d.ip.port))
			if err != nil {
				log.Println("Failed to connect:", err.Error())
				log.Println("Trying reset the connectiod...")
				time.Sleep(time.Millisecond * time.Duration(2000))
			} else {
				log.Println("Connected")
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
	}()
	return nil
}
