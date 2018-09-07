package tun

import "github.com/mysteriumnetwork/node/openvpn/config"

type LinuxTunnelSetup struct {
}

func (lts *LinuxTunnelSetup) Setup(config *config.GenericConfig) error {
	return nil
}

func (lts *LinuxTunnelSetup) Teardown() {

}

type GenericTunnelSetup struct {
}

func (gts *GenericTunnelSetup) Setup(config *config.GenericConfig) error {
	return nil
}

func (gts *GenericTunnelSetup) Teardown() {

}
