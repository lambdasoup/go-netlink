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

// Package netlink provides access to the Linux kernel's Netlink interface
package netlink

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"os"
	"syscall"

	"github.com/lambdasoup/go-netlink/log"
)

// from linux/netlink.h
var MSG_TYPES = map[MsgType]string{
	syscall.NLMSG_NOOP:    "NLMSG_NOOP",
	syscall.NLMSG_ERROR:   "NLMSG_ERROR",
	syscall.NLMSG_DONE:    "NLMSG_DONE",
	syscall.NLMSG_OVERRUN: "NLMSG_OVERRUN",
}

type MsgType uint16

func (t MsgType) String() string {
	return MSG_TYPES[t]
}

type NetlinkMsg struct {
	Len   uint32
	Type  MsgType
	Flags uint16
	Seq   uint32
	Pid   uint32
	Data  []byte
}

type NetlinkSocket struct {
	socketFd int
	lsa      *syscall.SockaddrNetlink
	seq      uint32
}

// Opens Netlink socket
func Open() (*NetlinkSocket, error) {
	// TODO remove Connector hardcode
	socketFd, err := syscall.Socket(syscall.AF_NETLINK, syscall.SOCK_DGRAM, syscall.NETLINK_CONNECTOR)
	if err != nil {
		return nil, err
	}
	lsa := &syscall.SockaddrNetlink{}
	lsa.Groups = 0
	lsa.Family = syscall.AF_NETLINK
	lsa.Pid = 0
	err = syscall.Bind(socketFd, lsa)
	return &NetlinkSocket{socketFd, lsa, 0xaffe}, err
}

// Closes the Connector
func (nls *NetlinkSocket) Close() {
	syscall.Close(nls.socketFd)
}

func (nls *NetlinkSocket) Send(data []byte) error {
	// TODO remove magic numbers
	msg := &NetlinkMsg{uint32(syscall.NLMSG_HDRLEN + len(data)), syscall.NLMSG_DONE, 0, nls.seq, uint32(os.Getpid()), data}
	nls.seq++

	log.Printf("\t\t\tNL SEND: %v", msg)

	// TODO remove magic number
	err := syscall.Sendto(nls.socketFd, msg.Bytes(), 0, nls.lsa)
	return err
}

func (msg *NetlinkMsg) Bytes() []byte {
	buf := new(bytes.Buffer)

	binary.Write(buf, binary.LittleEndian, msg.Len)
	binary.Write(buf, binary.LittleEndian, msg.Type)
	binary.Write(buf, binary.LittleEndian, msg.Flags)
	binary.Write(buf, binary.LittleEndian, msg.Seq)
	binary.Write(buf, binary.LittleEndian, msg.Pid)

	buf.Write(msg.Data)

	return buf.Bytes()
}

func (msg *NetlinkMsg) String() string {
	return fmt.Sprintf("NetlinkMsg{len: %d, %v, %x, seq: %d, port: %d, body: %d}", msg.Len, msg.Type, msg.Flags, msg.Seq, msg.Pid, len(msg.Data))
}

func (nls *NetlinkSocket) Receive() ([]byte, error) {
	// TODO remove magic numbers
	rb := make([]byte, 8192)
	_, _, err := syscall.Recvfrom(nls.socketFd, rb, 0)
	if err != nil {
		return nil, err
	}

	msg, err := parseNetlinkMsg(rb)
	log.Printf("\t\t\tNL RECV: %v", msg)
	return msg.Data, err
}

func parseNetlinkMsg(bs []byte) (*NetlinkMsg, error) {
	msg := &NetlinkMsg{}
	buf := bytes.NewBuffer(bs)

	err := error(nil)
	err = binary.Read(buf, binary.LittleEndian, &msg.Len)
	err = binary.Read(buf, binary.LittleEndian, &msg.Type)
	err = binary.Read(buf, binary.LittleEndian, &msg.Flags)
	err = binary.Read(buf, binary.LittleEndian, &msg.Seq)
	err = binary.Read(buf, binary.LittleEndian, &msg.Pid)

	msg.Data = make([]byte, msg.Len-syscall.NLMSG_HDRLEN)

	_, err = buf.Read(msg.Data)

	// check for truncated data
	for {
		bs := make([]byte, 1)
		_, eof := buf.Read(bs)
		if eof != nil {
			break
		}
		if bs[0] == 0 {
			continue
		}

		err = errors.New("NL parse left truncated data")
	}

	return msg, err
}
