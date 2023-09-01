//go:build windows

// Nebulant
// Copyright (C) 2023  Develatio Technologies S.L.

// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.

// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.

// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

package ipc

import (
	"net"

	"github.com/Microsoft/go-winio"
)

func (p *IPC) listen() (net.Listener, error) {
	path := `\\.\pipe\` + "ipc_" + p.uuid

	l, err := winio.ListenPipe(path, nil)
	if err != nil {
		return nil, err
	}
	return l, nil
}

func Read(ipsid string, ipcid string, msg string) (string, error) {
	path := `\\.\pipe\` + "ipc_" + ipsid
	c, err := winio.DialPipe(path, nil)
	if err != nil {
		return "", err
	}
	defer c.Close()

	_, err = c.Write([]byte(ipsid + " " + ipcid + " " + msg))
	if err != nil {
		return "", err
	}

	buf := make([]byte, 1024)

	n, err := c.Read(buf[:])
	if err != nil {
		return "", err
	}
	return string(buf[0:n]), nil
}
