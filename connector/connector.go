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

// Package connector provides access to the Connector subsystem via Netlink
package connector

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"

	"github.com/lambdasoup/go-netlink/log"
	"github.com/lambdasoup/go-netlink/netlink"
)

const (
	CNMSG_HDRLEN = 20
)

// From uapi/linux/connector.h
const (
	CN_W1_IDX = 3
	CN_W1_VAL = 1
)

// Response types
const (
	RESPONSE_TYPE_ECHO = iota
	RESPONSE_TYPE_REPLY
	RESPONSE_TYPE_UNRELATED
)

type CbId struct {
	Idx uint32
	Val uint32
}

type ConnectorMsg struct {
	id    CbId
	seq   uint32
	ack   uint32
	len   uint16
	flags uint16
	data  []byte
}

type Connector struct {
	nls *netlink.Socket
	id  CbId
	seq uint32
}

// Opens Connector
func Open(id CbId) (*Connector, error) {
	nls, err := netlink.Open()
	if err != nil {
		return nil, err
	}
	// TODO generate random sequence nr
	return &Connector{nls, id, 0xdead}, nil
}

func (c *ConnectorMsg) Data() []byte {
	return c.data
}

// Closes the Connector
func (c *Connector) Close() {
	c.nls.Close()
}

func (c *Connector) send(msg *ConnectorMsg) error {
	c.seq = c.seq + 1

	log.Printf("\t\tCN SEND: %v", msg)

	return c.nls.Send(msg.Bytes())
}

func (c *Connector) Receive(id *MsgId) (msg *ConnectorMsg, rtype int, err error) {
	data, err := c.nls.Receive()
	if err != nil {
		return
	}
	msg, err = parseConnectorMsg(data)
	if err != nil {
		return
	}

	if msg.ack == id.seq+1 {
		rtype = RESPONSE_TYPE_REPLY
	} else if msg.seq == id.seq {
		rtype = RESPONSE_TYPE_ECHO
	} else {
		rtype = RESPONSE_TYPE_UNRELATED
	}

	log.Printf("\t\tCN RECV: %v", msg)

	return
}

func (msg *ConnectorMsg) String() string {
	return fmt.Sprintf("ConnectorMsg{%v, seq: %d, ack: %d, data: %d, flags: %d}", msg.id, msg.seq, msg.ack, msg.len, msg.flags)
}

func (c *Connector) Request(req []byte) ([]byte, error) {
	id, err := c.Send(req)
	if err != nil {
		return nil, err
	}
	res, rtype, err := c.Receive(id)
	if err != nil {
		return nil, err
	}
	if rtype != RESPONSE_TYPE_REPLY {
		return nil, fmt.Errorf("unexpected response type %d", rtype)
	}

	return res.data, nil
}

type MsgId struct {
	seq uint32
}

func (c *Connector) Send(req []byte) (*MsgId, error) {
	// TODO remove magic numbers
	seq := c.seq
	msg := &ConnectorMsg{c.id, seq, 0, uint16(len(req)), 0, req}
	return &MsgId{seq}, c.send(msg)
}

func parseConnectorMsg(bs []byte) (*ConnectorMsg, error) {
	msg := &ConnectorMsg{}
	buf := bytes.NewBuffer(bs)

	err := error(nil)
	// TODO LE vs BE?
	err = binary.Read(buf, binary.LittleEndian, &msg.id)
	err = binary.Read(buf, binary.LittleEndian, &msg.seq)
	err = binary.Read(buf, binary.LittleEndian, &msg.ack)
	err = binary.Read(buf, binary.LittleEndian, &msg.len)
	err = binary.Read(buf, binary.LittleEndian, &msg.flags)

	msg.data = make([]byte, msg.len)

	n, err := buf.Read(msg.data)
	if err != nil {
		return nil, err
	}

	if n != int(msg.len) {
		return nil, errors.New("buffer size mismatch")
	}

	return msg, nil
}

func (msg *ConnectorMsg) Bytes() []byte {
	buf := new(bytes.Buffer)

	binary.Write(buf, binary.LittleEndian, msg.id.Idx)
	binary.Write(buf, binary.LittleEndian, msg.id.Val)
	binary.Write(buf, binary.LittleEndian, msg.seq)
	binary.Write(buf, binary.LittleEndian, msg.ack)
	binary.Write(buf, binary.LittleEndian, uint16(len(msg.data)))
	binary.Write(buf, binary.LittleEndian, msg.flags)

	buf.Write(msg.data)

	return buf.Bytes()
}
