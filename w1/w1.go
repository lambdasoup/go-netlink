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

// W1 is a 1-Wire connection
type W1 struct {
	c *connector.Connector
}

// ListMasters returns a list of the current list masters
func (w1 *W1) ListMasters() (masters []Master, err error) {
	log.Print("W1 LIST MASTERS")

	// send search request
	cmd := &msg{listMasters, 0, 0, nil, nil, 0, nil}
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

func (w1 *W1) request(req *msg, statusReplies int) (res *msg, err error) {
	log.Printf("\tW1 REQUEST: %v", req)

	msgID, err := w1.c.Send(req.toBytes())
	if err != nil {
		return
	}

	// we need to await all status replies and the actual response
	// these are all out-of-order
	for statusReplies > 0 || res == nil {
		data, rtype, err := w1.c.Receive(msgID)
		if err != nil {
			return nil, err
		}
		msg := parseW1Msg(data)
		switch rtype {
		case connector.ResponseTypeReply:
			log.Printf("\tW1 RECV REPLY: %v", msg)
			res = msg
		case connector.ResponseTypeEcho:
			log.Printf("\tW1 RECV STATUS: %v", msg)
			if msg.status != 0 {
				err = fmt.Errorf("status error %d", msg.status)
				return nil, err
			}
			statusReplies--
		case connector.ResponseTypeUnrelated:
			err = errors.New("received unexpected unrelated response")
			return nil, err
		}
	}

	log.Printf("\tW1 RECV: %v", res)
	return
}

func (w1 *W1) send(req *msg) (err error) {
	log.Printf("\tW1 SEND: %v", req)

	msgID, err := w1.c.Send(req.toBytes())
	if err != nil {
		return
	}

	data, rtype, err := w1.c.Receive(msgID)
	if err != nil {
		return
	}

	msg := parseW1Msg(data)
	switch rtype {
	case connector.ResponseTypeReply:
		return errors.New("received unexpected request response")
	case connector.ResponseTypeEcho:
		log.Printf("\tW1 RECV STATUS: %v", msg)
		if msg.status != 0 {
			err = fmt.Errorf("status error %d", msg.status)
			return err
		}
		return nil
	case connector.ResponseTypeUnrelated:
		return errors.New("received unexpected unrelated response")
	}

	panic(fmt.Sprintf("unexpected connector msg type %d", rtype))
}

// Open a connection to the 1-Wire subsystem
func (w1 *W1) Open() (err error) {
	c, err := connector.Open(connector.W1)
	w1.c = c
	return
}

// Close closes this 1-Wire connection
func (w1 *W1) Close() {
	w1.c.Close()
}
