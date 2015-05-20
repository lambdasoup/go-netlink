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

	"github.com/lambdasoup/go-netlink/log"
)

// Master is a 1-Wire Master device
type Master struct {
	id uint32
	w1 *W1
}

// Close this Slave's connection
func (ms *Master) Close() {
	ms.w1.Close()
}

// ListSlaves returns a list of this master's slaves
func (ms *Master) ListSlaves() (slaves []Slave, err error) {
	log.Print("W1 LIST SLAVES")

	// send list slaves request
	c := cmd{cmdListSlaves, 0, nil}
	req := &msg{masterCmd, 0, uint16(len(c.toBytes())), ms, nil, 0, c.toBytes()}

	msg, err := ms.w1.request(req, 1)
	if err != nil {
		return
	}
	buf := bytes.NewBuffer(msg.data)

	// skip W1_CMD part
	buf.Next(4)
	for i := 0; i < int(msg.len-4); i = i + 8 {
		slave := new(Slave)
		binary.Read(buf, binary.LittleEndian, &slave.family)
		binary.Read(buf, binary.LittleEndian, &slave.uid)
		binary.Read(buf, binary.LittleEndian, &slave.crc)
		slave.master = ms
		slaves = append(slaves, *slave)
	}

	return
}

func (ms *Master) readSlave(slave *Slave, args []byte, count int) (data []byte, err error) {
	log.Print("W1 READ SLAVE")

	cmdW := cmd{cmdWrite, 0, args}
	cmdR := cmd{cmdRead, 0, make([]byte, count)}

	buf := bytes.NewBuffer(make([]byte, 0))
	buf.Write(cmdW.toBytes())
	buf.Write(cmdR.toBytes())
	cmd := buf.Bytes()

	req := &msg{slaveCmd, 0, uint16(len(cmd)), nil, slave, 0, cmd}

	msg, err := ms.w1.request(req, 2)
	if err != nil {
		return
	}
	// throw away w1 cmd header
	data = msg.data[4:]

	return
}

func (ms *Master) writeSlave(slave *Slave, args []byte) (err error) {
	log.Print("W1 WRITE SLAVE")

	cmd := cmd{cmdWrite, 0, args}
	req := &msg{slaveCmd, 0, uint16(len(cmd.toBytes())), nil, slave, 0, cmd.toBytes()}

	err = ms.w1.send(req)
	return
}
