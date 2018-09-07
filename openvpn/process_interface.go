package openvpn

// Process defines openvpn process interface with basic controls
// It must die!
type Process interface {
	Start() error
	Wait() error
	Stop()
}
