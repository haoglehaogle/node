/*
 * Copyright (C) 2017 The "MysteriumNetwork/node" Authors.
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program.  If not, see <http://www.gnu.org/licenses/>.
 */

package extcmd

import (
	"errors"
	"github.com/mysteriumnetwork/node/openvpn/config"
	"github.com/mysteriumnetwork/node/openvpn/extcmd/cmd"
	"github.com/mysteriumnetwork/node/openvpn/extcmd/management"
	"github.com/mysteriumnetwork/node/openvpn/extcmd/tun"
	"sync"
	"time"
)

type tunnelSetup interface {
	Setup(config *config.GenericConfig) error
	Teardown()
}

type OpenvpnProcess struct {
	tunSetup   tunnelSetup
	config     *config.GenericConfig
	management *management.Management
	cmd        *cmd.Wrapper
}

func NewOpenvpnProcess(openvpnBinary string, config *config.GenericConfig, middlewares ...management.Middleware) *OpenvpnProcess {
	return &OpenvpnProcess{
		tunSetup:   tun.NewSetup(),
		config:     config,
		management: management.NewManagement(management.LocalhostOnRandomPort, "[client-management] ", middlewares...),
		cmd:        cmd.NewWrapper(openvpnBinary, "[openvpn-process] "),
	}
}

func (openvpn *OpenvpnProcess) Start() error {
	if err := openvpn.tunSetup.Setup(openvpn.config); err != nil {
		return err
	}

	err := openvpn.management.WaitForConnection()
	if err != nil {
		return err
	}

	addr := openvpn.management.BoundAddress
	openvpn.config.SetManagementAddress(addr.IP, addr.Port)

	// Fetch the current arguments
	arguments, err := (*openvpn.config).ToArguments()
	if err != nil {
		return err
	}

	//nil returned from process.Start doesn't guarantee that openvpn itself initialized correctly and accepted all arguments
	//it simply means that OS started process with specified args
	err = openvpn.cmd.Start(arguments)
	if err != nil {
		openvpn.management.Stop()
		return err
	}

	select {
	case connAccepted := <-openvpn.management.Connected:
		if connAccepted {
			return nil
		}
		return errors.New("management failed to accept connection")
	case exitError := <-openvpn.cmd.CmdExitError:
		openvpn.management.Stop()
		if exitError != nil {
			return exitError
		}
		return errors.New("openvpn process died too early")
	case <-time.After(2 * time.Second):
		return errors.New("management connection wait timeout")
	}
}

func (openvpn *OpenvpnProcess) Wait() error {
	return openvpn.cmd.Wait()
}

func (openvpn *OpenvpnProcess) Stop() {
	waiter := sync.WaitGroup{}
	//TODO which to signal for close first ?
	//if we stop process before management, managemnt won't have a chance to send any commands from middlewares on stop
	//if we stop management first - it will miss important EXITING state from process
	waiter.Add(1)
	go func() {
		defer waiter.Done()
		openvpn.cmd.Stop()
	}()

	waiter.Add(1)
	go func() {
		defer waiter.Done()
		openvpn.management.Stop()
	}()

	waiter.Wait()
}
