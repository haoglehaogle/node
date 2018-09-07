// +build !linux

package tun

func NewSetup() *GenericTunnelSetup {
	return &GenericTunnelSetup{}
}
