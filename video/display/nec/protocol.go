package nec

import "strconv"

// SetSource sets source s on destination d
// d is ignored in this device
func (d *Display) SetSource(s int, dest int) {
	d.conn.Write([]byte(strconv.Itoa(s) + "!"))
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
