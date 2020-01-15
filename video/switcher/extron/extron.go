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

type extron struct {
	conn        net.Conn
	rxLineCount int
	ipHost      string
	ipPort      int
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

// sendSpecial is a wrapper for send which adds escape chars as required by some commands
func (e *extron) sendSpecial(s string) {
	e.send("\x1B" + s + "\x0D")
}

// send pushed command to the switcher
func (e *extron) send(s string) {
	e.conn.Write([]byte(s))
	s = strconv.QuoteToASCII(s)
	// log.Print("Tx::", s[1:len(s)-1])
}

// SetSource sets source s on destination d
// d is ignored in this device
func (e *extron) SetSource(s int, d int) {
	e.conn.Write([]byte(strconv.Itoa(s) + "!"))
}

// GetSource returns current input number
func (e *extron) GetSource() int {
	return e.Input
}

// SetPower is unsued on this device
func (e *extron) SetPower(p bool) {}

// Getpower always returns true on this device
func (e *extron) GetPower() bool { return true }

// New creates a new instance, and initialises communication
func (e *extron) Init(ip string) error {

	// Default Port
	e.ipPort = 23

	// Store Host & Port
	for i, x := range strings.Split(ip, `:`) {
		switch i {
		case 0:
			e.ipHost = x
		case 1:
			e.ipPort, _ = strconv.Atoi(x)
		}
	}

	go func() {
		for {
			var err error
			e.conn, err = net.Dial("tcp", e.ipHost+`:`+strconv.Itoa(+e.ipPort))
			if err != nil {
				fmt.Println("Failed to connect:", err.Error())
				fmt.Println("Trying reset the connection...")
				time.Sleep(time.Millisecond * time.Duration(2000))
			} else {
				// Create new Reader
				r := bufio.NewReader(e.conn)
				e.rxLineCount = 0
				for {
					message, err := r.ReadString('\n')
					if err != nil {
						log.Println("RxErr::", err)
						break
					}
					log.Print("Rx::", message)
					e.rxLineCount++ // increment line count

					// Init Device
					if e.rxLineCount == 3 {
						e.sendSpecial("3CV") // Set Verbose Mode
						e.send("N")          // Request Part Number
						e.send("1I")         // Query Model Name
						e.send("2I")         // Query Model Description
						e.send("Q")          // Query firmware version
						e.send("!")          // Query Input Number
					}

					// Process Feedback
					message = strings.TrimSpace(message)
					switch {
					case strings.HasPrefix(message, "Pno"):
						e.PartNumber = strings.TrimPrefix(message, "Pno")
						// Request Additional Feedback
						switch e.PartNumber {
						case "60-1603-01", "60-1604-01", "60-1605-01", "60-1606-01":
							e.sendSpecial("LS") // Request status of all signals

						case "60-1238-51":
							e.sendSpecial("0LS") // Request status of all signals
						}
					case strings.HasPrefix(message, "Info01*"):
						e.ModelName = strings.TrimPrefix(message, "Info01*")
					case strings.HasPrefix(message, "Info02*"):
						e.ModelDesc = strings.TrimPrefix(message, "Info02*")
					case strings.HasPrefix(message, "Info02*"):
						e.FirmwareVer = strings.TrimPrefix(message, "Ver01*")
					case strings.HasPrefix(message, "Sig"):
						x := strings.TrimPrefix(message, "Sig")
						// Pull out the output Status
						y := strings.Split(x, `*`)
						e.SignalOutput, _ = strconv.ParseBool(y[1])
						// Pull out the Input Status
						z := strings.Split(y[0], ` `)
						for i, s := range z {
							if len(e.SignalInput) <= i {
								e.SignalInput = append(e.SignalInput, false)
							}
							e.SignalInput[i], _ = strconv.ParseBool(s)
						}
					}
				}
			}
		}
	}()
	return nil
}
