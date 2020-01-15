package extron

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"strconv"
	"strings"
	"time"
)

// Device is a representation of an Extron Switcher
type Device struct {
	// Communications
	conn net.Conn
	ip   struct {
		connected bool
		host      string
		port      int
	}
	// MetaData
	PartNumber  string
	ModelName   string
	ModelDesc   string
	FirmwareVer string
	// Status
	Input     int
	VideoMute bool
	signal    struct {
		input  []bool
		output bool
	}
	// Misc
	rxLineCount int
}

// sendSpecial is a wrapper for send which adds escape chars as required by some commands
func (d *Device) sendSpecial(s string) {
	d.send("\x1B" + s + "\x0D")
}

// send pushed command to the switcher
func (d *Device) send(s string) {
	d.conn.Write([]byte(s))
	s = strconv.QuoteToASCII(s)
	// log.Print("Tx::", s[1:len(s)-1])
}

// GetHost returns current host
func (d Device) GetHost() string {
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
func (d *Device) SetPower(p bool) {}

// GetPower always returns true on this device
func (d *Device) GetPower() bool { return true }

// Init sets defaults, and initialises communication
func (d *Device) Init(ip string) error {

	// Default Port
	d.ip.port = 23

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
				fmt.Println("Failed to connect:", err.Error())
				fmt.Println("Trying reset the connection...")
				time.Sleep(time.Millisecond * time.Duration(2000))
			} else {
				// Create new Reader
				r := bufio.NewReader(d.conn)
				d.rxLineCount = 0
				for {
					message, err := r.ReadString('\n')
					if err != nil {
						log.Println("RxErr::", err)
						break
					}
					log.Print("Rx::", message)
					d.rxLineCount++ // increment line count

					// Init Device
					if d.rxLineCount == 3 {
						d.sendSpecial("3CV") // Set Verbose Mode
						d.send("N")          // Request Part Number
						d.send("1I")         // Query Model Name
						d.send("2I")         // Query Model Description
						d.send("Q")          // Query firmware version
						d.send("!")          // Query Input Number
					}

					// Process Feedback
					message = strings.TrimSpace(message)
					switch {
					case strings.HasPrefix(message, "Pno"):
						d.PartNumber = strings.TrimPrefix(message, "Pno")
						// Request Additional Feedback
						switch d.PartNumber {
						case "60-1603-01", "60-1604-01", "60-1605-01", "60-1606-01":
							d.sendSpecial("LS") // Request status of all signals

						case "60-1238-51":
							d.sendSpecial("0LS") // Request status of all signals
						}
					case strings.HasPrefix(message, "Info01*"):
						d.ModelName = strings.TrimPrefix(message, "Info01*")
					case strings.HasPrefix(message, "Info02*"):
						d.ModelDesc = strings.TrimPrefix(message, "Info02*")
					case strings.HasPrefix(message, "Info02*"):
						d.FirmwareVer = strings.TrimPrefix(message, "Ver01*")
					case strings.HasPrefix(message, "Sig"):
						x := strings.TrimPrefix(message, "Sig")
						// Pull out the output Status
						y := strings.Split(x, `*`)
						d.signal.output, _ = strconv.ParseBool(y[1])
						// Pull out the Input Status
						z := strings.Split(y[0], ` `)
						for i, s := range z {
							if len(d.signal.input) <= i {
								d.signal.input = append(d.signal.input, false)
							}
							d.signal.input[i], _ = strconv.ParseBool(s)
						}
					}
				}
			}
		}
	}()
	return nil
}
