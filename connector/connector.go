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

type msg struct {
	id    CbId
	seq   uint32
	ack   uint32
	len   uint16
	flags uint16
	data  []byte
}

// Connector is a Linux Connector
type Connector struct {
	nls *netlink.Socket
	id  CbId
	seq uint32
}

// Open a new Connector
func Open(id CbId) (*Connector, error) {
	nls, err := netlink.Open()
	if err != nil {
		return nil, err
	}
	// TODO generate random sequence nr
	return &Connector{nls, id, 0xdead}, nil
}

// Close the Connector
func (c *Connector) Close() {
	c.nls.Close()
}

func (c *Connector) send(m *msg) error {
	c.seq = c.seq + 1

	log.Printf("\t\tCN SEND: %v", m)

	return c.nls.Send(m.bytes())
}

// Receive data on this Connector
func (c *Connector) Receive(id *MsgId) (body []byte, rtype int, err error) {
	data, err := c.nls.Receive()
	if err != nil {
		return
	}
	m, err := parseConnectorMsg(data)
	if err != nil {
		return
	}
	body = m.data

	if m.ack == id.seq+1 {
		rtype = RESPONSE_TYPE_REPLY
	} else if m.seq == id.seq {
		rtype = RESPONSE_TYPE_ECHO
	} else {
		rtype = RESPONSE_TYPE_UNRELATED
	}

	log.Printf("\t\tCN RECV: %v", m)

	return
}

func (m *msg) String() string {
	return fmt.Sprintf("ConnectorMsg{%v, seq: %d, ack: %d, data: %d, flags: %d}", m.id, m.seq, m.ack, m.len, m.flags)
}

// Request data on this Connector
func (c *Connector) Request(req []byte) ([]byte, error) {
	id, err := c.Send(req)
	if err != nil {
		return nil, err
	}
	body, rtype, err := c.Receive(id)
	if err != nil {
		return nil, err
	}
	if rtype != RESPONSE_TYPE_REPLY {
		return nil, fmt.Errorf("unexpected response type %d", rtype)
	}

	return body, nil
}

// MsgId identifies messages
type MsgId struct {
	seq uint32
}

// Send data on this Connector
func (c *Connector) Send(req []byte) (*MsgId, error) {
	// TODO remove magic numbers
	seq := c.seq
	m := &msg{c.id, seq, 0, uint16(len(req)), 0, req}
	return &MsgId{seq}, c.send(m)
}

func parseConnectorMsg(bs []byte) (*msg, error) {
	m := &msg{}
	buf := bytes.NewBuffer(bs)

	err := error(nil)
	// TODO LE vs BE?
	err = binary.Read(buf, binary.LittleEndian, &m.id)
	err = binary.Read(buf, binary.LittleEndian, &m.seq)
	err = binary.Read(buf, binary.LittleEndian, &m.ack)
	err = binary.Read(buf, binary.LittleEndian, &m.len)
	err = binary.Read(buf, binary.LittleEndian, &m.flags)

	m.data = make([]byte, m.len)

	n, err := buf.Read(m.data)
	if err != nil {
		return nil, err
	}

	if n != int(m.len) {
		return nil, errors.New("buffer size mismatch")
	}

	return m, nil
}

func (m *msg) bytes() []byte {
	buf := new(bytes.Buffer)

	binary.Write(buf, binary.LittleEndian, m.id.Idx)
	binary.Write(buf, binary.LittleEndian, m.id.Val)
	binary.Write(buf, binary.LittleEndian, m.seq)
	binary.Write(buf, binary.LittleEndian, m.ack)
	binary.Write(buf, binary.LittleEndian, uint16(len(m.data)))
	binary.Write(buf, binary.LittleEndian, m.flags)

	buf.Write(m.data)

	return buf.Bytes()
}
