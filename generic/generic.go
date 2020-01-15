package generic

// Device represents the base control of all devices
type Device interface {
	SetSource(s int, dest int)
	GetSource() int
	SetPower(p bool)
	GetPower() bool
	Init(ip string) error
	GetHost() string
}
