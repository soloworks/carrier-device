package samsung

import "bufio"

// SetSource sets source s on destination d
// d is ignored in this device
func (d *Display) SetSource(s int, dest int) {

}

// SetPower is unsued on this device
func (d *Display) SetPower(p bool) {
	switch p {
	case true:
		err := d.conn.Send(d.packMessage(0x11, 0x01)) // Power On
		if err != nil {
			log.Error(err)
		}
	case false:
		err := d.conn.Send(d.packMessage(0x11, 0x00)) // Power Off
		if err != nil {
			log.Error(err)
		}
	}
}

func (d *Display) packMessage(cmd byte, data ...byte) []byte {

	// Build the message
	var b []byte                   // Create Array
	b = append(b, cmd)             // Add Command
	b = append(b, byte(d.id))      // Add ID
	b = append(b, byte(len(data))) // Add Data Length
	b = append(b, data...)         // Add Data
	b = append(b, chksum(b))       // Add Checksum
	b = append([]byte{0xAA}, b...) // Add Header
	return b
}

func chksum(msg []byte) byte {
	var chk int
	for _, b := range msg {
		chk += int(b)
	}
	return byte(chk)
}

func (d *Display) poll() {
	d.conn.Poll(d.packMessage(0x00)) // Query Status
	d.conn.Poll(d.packMessage(0x11)) // Query Power State
	d.conn.Poll(d.packMessage(0x12)) // Query Audio Gain
	d.conn.Poll(d.packMessage(0x13)) // Query Audio Mute
	d.conn.Poll(d.packMessage(0x14)) // Query Input Source
	d.conn.Poll(d.packMessage(0x0B)) // Query Serial Number
	d.conn.Poll(d.packMessage(0x0E)) // Query Firmware Version
	d.conn.Poll(d.packMessage(0x10)) // Query Model Number
}

// Decode incoming bytes and return each recognised packet
func (d *Display) decoder(r *bufio.Reader) ([]byte, error) {

	// Peek at first 4 of buffer
	p, err := r.Peek(4)
	// Not enough to peek, leave it for now
	if err != nil {
		return nil, nil
	}
	// Check for STX, get rid of a byte of garbage
	if p[0] != 0xAA {
		log.Error("Found Garbage byte in Response")
		r.Discard(1)
		return nil, nil
	}
	// Calculate length of expected packet based on data len field
	pktLen := 5 + int(p[3])
	// Check if enough data is in the buffer
	p, err = r.Peek(pktLen)
	// Not enough to peek, leave it for now
	if err != nil {
		return nil, nil
	}
	r.Discard(pktLen)

	// Verify Checksum is correct
	if chksum(p[1:len(p)-1]) != p[len(p)-1] {
		log.Error("Found corrupt message")
		return nil, nil
	}
	return p, nil
}

// Pass data through as is encoded already
func (d *Display) encoder(b []byte) ([]byte, error) {
	return b, nil
}
