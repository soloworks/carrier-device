package samsung

// SetSource sets source s on destination d
// d is ignored in this device
func (d *Display) SetSource(s int, dest int) {

}

// SetPower is unsued on this device
func (d *Display) SetPower(p bool) {
	switch p {
	case true:
		d.Send(0x11, 0x01) // Power On
	case false:
		d.Send(0x11, 0x00) // Power Off
	}
}

func packMessage(id int, cmd byte, data ...byte) []byte {

	// Build the message
	var b []byte                   // Create Array
	b = append(b, cmd)             // Add Command
	b = append(b, byte(id))        // Add ID
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
