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

// Package w1 provides access to One-Wire devices over Netlink / Connector
package w1

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"

	"github.com/lambdasoup/go-netlink/connector"
	"github.com/lambdasoup/go-netlink/log"
)

type w1CmdType uint8

// From drivers/w1/w1_netlink.h
const (
	W1_CMD_READ w1CmdType = iota
	W1_CMD_WRITE
	W1_CMD_SEARCH
	W1_CMD_ALARM_SEARCH
	W1_CMD_TOUCH
	W1_CMD_RESET
	W1_CMD_SLAVE_ADD
	W1_CMD_SLAVE_REMOVE
	W1_CMD_LIST_SLAVES
)

type w1MsgType uint8

// One-Wire command types
const (
	W1_SLAVE_ADD w1MsgType = iota
	W1_SLAVE_REMOVE
	W1_MASTER_ADD
	W1_MASTER_REMOVE
	W1_MASTER_CMD
	W1_SLAVE_CMD
	W1_LIST_MASTERS
)

// W1Cmd is a One-Wire command
// From drivers/w1/w1_netlink.h
type W1Cmd struct {
	cmd  w1CmdType
	res  uint8
	data []byte
}

func (cmd *W1Cmd) String() string {
	return fmt.Sprintf("W1Cmd{%v, data %x}", cmd.cmd, cmd.data)
}

// W1Cmd is a One-Wire message
// From drivers/w1/w1_netlink.h
type W1Msg struct {
	w1Type w1MsgType
	status uint8
	len    uint16
	master *Master
	slave  *Slave
	res    uint32
	data   []byte
}

func (msg *W1Msg) String() string {
	return fmt.Sprintf("W1Msg{%v, status %v, len %v, master %v, slave %v, data %x}", msg.w1Type, msg.status, len(msg.data), msg.master, msg.slave, msg.data)
}

type W1 struct {
	c *connector.Connector
}

// ListMasters returns a list of the current list masters
func (w1 *W1) ListMasters() (masters []Master, err error) {
	log.Print("W1 LIST MASTERS")

	// send search request
	cmd := &W1Msg{W1_LIST_MASTERS, 0, 0, nil, nil, 0, nil}
	msg, err := w1.request(cmd, -1)
	if err != nil {
		return
	}
	buf := bytes.NewBuffer(msg.data)
	for i := 0; i < int(msg.len); i = i + 4 {
		var id uint32
		binary.Read(buf, binary.LittleEndian, &id)
		masters = append(masters, Master{id, w1})
	}
	return
}

func (w1 *W1) request(req *W1Msg, statusReplies int) (res *W1Msg, err error) {
	log.Printf("\tW1 REQUEST: %v", req)

	msgId, err := w1.c.Send(req.toBytes())
	if err != nil {
		return
	}

	// we need to await all status replies and the actual response
	// these are all out-of-order
	for statusReplies > 0 || res == nil {
		data, rtype, err := w1.c.Receive(msgId)
		if err != nil {
			return nil, err
		}
		msg := parseW1Msg(data)
		switch rtype {
		case connector.RESPONSE_TYPE_REPLY:
			log.Printf("\tW1 RECV REPLY: %v", msg)
			res = msg
		case connector.RESPONSE_TYPE_ECHO:
			log.Printf("\tW1 RECV STATUS: %v", msg)
			if msg.status != 0 {
				err = fmt.Errorf("status error %d", msg.status)
				return nil, err
			}
			statusReplies--
		case connector.RESPONSE_TYPE_UNRELATED:
			err = errors.New("received unexpected unrelated response")
			return nil, err
		}
	}

	log.Printf("\tW1 RECV: %v", res)
	return
}

func (w1 *W1) send(req *W1Msg) (err error) {
	log.Printf("\tW1 SEND: %v", req)

	msgId, err := w1.c.Send(req.toBytes())
	if err != nil {
		return
	}

	data, rtype, err := w1.c.Receive(msgId)
	if err != nil {
		return
	}

	msg := parseW1Msg(data)
	switch rtype {
	case connector.RESPONSE_TYPE_REPLY:
		return errors.New("received unexpected request response")
	case connector.RESPONSE_TYPE_ECHO:
		log.Printf("\tW1 RECV STATUS: %v", msg)
		if msg.status != 0 {
			err = fmt.Errorf("status error %d", msg.status)
			return err
		}
		return nil
	case connector.RESPONSE_TYPE_UNRELATED:
		return errors.New("received unexpected unrelated response")
	}

	panic(fmt.Sprintf("unexpected connector msg type %d", rtype))
}

// Open opens a connection to the 1-Wire subsystem
func (w1 *W1) Open() (err error) {
	c, err := connector.Open(connector.CbId{connector.CN_W1_IDX, connector.CN_W1_VAL})
	w1.c = c
	return
}

// Close closes this 1-Wire connection
func (w1 *W1) Close() {
	w1.c.Close()
}

func (cmd *W1Cmd) toBytes() []byte {
	buf := new(bytes.Buffer)

	binary.Write(buf, binary.LittleEndian, cmd.cmd)
	binary.Write(buf, binary.LittleEndian, cmd.res)
	binary.Write(buf, binary.LittleEndian, uint16(len(cmd.data)))
	buf.Write(cmd.data)

	return buf.Bytes()
}

func parseW1Msg(bs []byte) *W1Msg {
	msg := &W1Msg{}
	buf := bytes.NewBuffer(bs)

	binary.Read(buf, binary.LittleEndian, &msg.w1Type)
	binary.Read(buf, binary.LittleEndian, &msg.status)
	binary.Read(buf, binary.LittleEndian, &msg.len)

	// master or slave id depending on msg type
	switch msg.w1Type {
	case W1_SLAVE_ADD, W1_SLAVE_REMOVE, W1_SLAVE_CMD:
		msg.slave = new(Slave)
		binary.Read(buf, binary.LittleEndian, &msg.slave.family)
		binary.Read(buf, binary.LittleEndian, &msg.slave.uid)
		binary.Read(buf, binary.LittleEndian, &msg.slave.crc)

	case W1_MASTER_ADD, W1_MASTER_REMOVE, W1_MASTER_CMD:
		msg.master = new(Master)
		binary.Read(buf, binary.LittleEndian, &msg.master.id)
		buf.Next(4)

	case W1_LIST_MASTERS:
		buf.Next(8)
	}

	msg.data = make([]byte, msg.len)
	buf.Read(msg.data)

	return msg
}

func (msg *W1Msg) toBytes() []byte {
	buf := new(bytes.Buffer)

	binary.Write(buf, binary.LittleEndian, msg.w1Type)
	binary.Write(buf, binary.LittleEndian, msg.status)
	binary.Write(buf, binary.LittleEndian, msg.len)
	// some messages do not have a master set
	if msg.master != nil {
		binary.Write(buf, binary.LittleEndian, msg.master.id)
		binary.Write(buf, binary.LittleEndian, msg.res)
	} else if msg.slave != nil {
		binary.Write(buf, binary.LittleEndian, msg.slave.family)
		binary.Write(buf, binary.LittleEndian, msg.slave.uid)
		binary.Write(buf, binary.LittleEndian, msg.slave.crc)
	} else {
		binary.Write(buf, binary.LittleEndian, uint32(0))
		binary.Write(buf, binary.LittleEndian, msg.res)
	}

	buf.Write(msg.data)

	return buf.Bytes()
}

type Master struct {
	id uint32
	w1 *W1
}

func (m *Master) Close() {
	m.w1.Close()
}

// ListSlaves returns a list of this master's slaves
func (m *Master) ListSlaves() (slaves []Slave, err error) {
	log.Print("W1 LIST SLAVES")

	// send list slaves request
	cmd := W1Cmd{W1_CMD_LIST_SLAVES, 0, nil}
	req := &W1Msg{W1_MASTER_CMD, 0, uint16(len(cmd.toBytes())), m, nil, 0, cmd.toBytes()}

	msg, err := m.w1.request(req, 1)
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
		slave.master = m
		slaves = append(slaves, *slave)
	}

	return
}

func (m *Master) readSlave(slave *Slave, args []byte, count int) (data []byte, err error) {
	log.Print("W1 READ SLAVE")

	cmdW := W1Cmd{W1_CMD_WRITE, 0, args}
	cmdR := W1Cmd{W1_CMD_READ, 0, make([]byte, count)}

	buf := bytes.NewBuffer(make([]byte, 0))
	buf.Write(cmdW.toBytes())
	buf.Write(cmdR.toBytes())
	cmd := buf.Bytes()

	req := &W1Msg{W1_SLAVE_CMD, 0, uint16(len(cmd)), nil, slave, 0, cmd}

	msg, err := m.w1.request(req, 2)
	if err != nil {
		return
	}
	// throw away w1 cmd header
	data = msg.data[4:]

	return
}

func (m *Master) writeSlave(slave *Slave, args []byte) (err error) {
	log.Print("W1 WRITE SLAVE")

	cmd := W1Cmd{W1_CMD_WRITE, 0, args}
	req := &W1Msg{W1_SLAVE_CMD, 0, uint16(len(cmd.toBytes())), nil, slave, 0, cmd.toBytes()}

	err = m.w1.send(req)
	return
}

type Slave struct {
	family byte
	uid    [6]byte
	crc    byte
	master *Master
}

func (s *Slave) Close() {
	s.master.Close()
}

func (s *Slave) IsFamily(family byte) bool {
	return s.family == family
}

func (s *Slave) String() string {
	return fmt.Sprintf("Slave{Family:%x, UID:%x}", s.family, s.uid)
}

func (s *Slave) Read(data []byte, count int) ([]byte, error) {
	return s.master.readSlave(s, data, count)
}

func (s *Slave) Write(data []byte) error {
	return s.master.writeSlave(s, data)
}
