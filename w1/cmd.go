// This file is part of go-netlink.
//
// Copyright (C) 2015 Max Hille <mh@lambdasoup.com>
//
// go-netlink is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// at your option) any later version.
//
// go-netlink is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with go-netlink.  If not, see <http://www.gnu.org/licenses/>.

package w1

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

type cmdType uint8

// From drivers/w1/w1_netlink.h
const (
	cmdRead cmdType = iota
	cmdWrite
	cmdSearch
	cmdAlarmSearch
	cmdTouch
	cmdReset
	cmdSlaveAdd
	cmdSlaveRemove
	cmdListSlaves
)

// Cmd is a 1-Wire command
// From drivers/w1/w1_netlink.h
type cmd struct {
	cmd  cmdType
	res  uint8
	data []byte
}

func (c *cmd) String() string {
	return fmt.Sprintf("W1Cmd{%v, data %x}", c.cmd, c.data)
}

func (c *cmd) toBytes() []byte {
	buf := new(bytes.Buffer)

	binary.Write(buf, binary.LittleEndian, c.cmd)
	binary.Write(buf, binary.LittleEndian, c.res)
	binary.Write(buf, binary.LittleEndian, uint16(len(c.data)))
	buf.Write(c.data)

	return buf.Bytes()
}
