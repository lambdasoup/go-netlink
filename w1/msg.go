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

type msgType uint8

// One-Wire command types
const (
	slaveAdd msgType = iota
	slaveRemove
	masterAdd
	masterRemove
	masterCmd
	slaveCmd
	listMasters
)

// Msg is a 1-Wire message
// From drivers/w1/w1_netlink.h
type msg struct {
	w1Type msgType
	status uint8
	len    uint16
	master *Master
	slave  *Slave
	res    uint32
	data   []byte
}

func (m *msg) String() string {
	return fmt.Sprintf("W1Msg{%v, status %v, len %v, master %v, slave %v, data %x}",
		m.w1Type, m.status, len(m.data), m.master, m.slave, m.data)
}

func parseW1Msg(bs []byte) *msg {
	msg := &msg{}
	buf := bytes.NewBuffer(bs)

	binary.Read(buf, binary.LittleEndian, &msg.w1Type)
	binary.Read(buf, binary.LittleEndian, &msg.status)
	binary.Read(buf, binary.LittleEndian, &msg.len)

	// master or slave id depending on msg type
	switch msg.w1Type {
	case slaveAdd, slaveRemove, slaveCmd:
		msg.slave = new(Slave)
		binary.Read(buf, binary.LittleEndian, &msg.slave.family)
		binary.Read(buf, binary.LittleEndian, &msg.slave.uid)
		binary.Read(buf, binary.LittleEndian, &msg.slave.crc)

	case masterAdd, masterRemove, masterCmd:
		msg.master = new(Master)
		binary.Read(buf, binary.LittleEndian, &msg.master.id)
		buf.Next(4)

	case listMasters:
		buf.Next(8)
	}

	msg.data = make([]byte, msg.len)
	buf.Read(msg.data)

	return msg
}

func (m *msg) toBytes() []byte {
	buf := new(bytes.Buffer)

	binary.Write(buf, binary.LittleEndian, m.w1Type)
	binary.Write(buf, binary.LittleEndian, m.status)
	binary.Write(buf, binary.LittleEndian, m.len)
	// some messages do not have a master set
	if m.master != nil {
		binary.Write(buf, binary.LittleEndian, m.master.id)
		binary.Write(buf, binary.LittleEndian, m.res)
	} else if m.slave != nil {
		binary.Write(buf, binary.LittleEndian, m.slave.family)
		binary.Write(buf, binary.LittleEndian, m.slave.uid)
		binary.Write(buf, binary.LittleEndian, m.slave.crc)
	} else {
		binary.Write(buf, binary.LittleEndian, uint32(0))
		binary.Write(buf, binary.LittleEndian, m.res)
	}

	buf.Write(m.data)

	return buf.Bytes()
}
