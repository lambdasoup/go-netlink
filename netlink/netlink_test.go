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

package netlink

import (
	"syscall"
	"testing"
)

func TestParseNetlinkMessage(t *testing.T) {
	var bs []byte

	// length
	bs = append(bs, 47, 0, 0, 0)
	// type, flags
	bs = append(bs, 3, 0, 0, 0)
	// seq, pid
	bs = append(bs, 57, 48, 0, 0, 0, 0, 0, 0)

	// payload
	bs = append(bs, 14, 0, 0, 0, 86, 4, 0, 0, 57, 48, 0, 0)
	bs = append(bs, 58, 48, 0, 0, 11, 0, 0, 0)
	bs = append(bs, 116, 101, 115, 116, 32, 114, 101, 112, 108, 121)

	msg, _ := parseNetlinkMsg(bs)

	assert(t, msg.len == uint32(47))
	assert(t, msg.msgType == syscall.NLMSG_DONE)
	assert(t, msg.flags == uint16(0))
	assert(t, msg.seq == uint32(12345))
	assert(t, msg.pid == uint32(0))

}

func TestBytes(t *testing.T) {
	var data []byte

	msg := &netlinkMsg{uint32(syscall.NLMSG_HDRLEN + len(data)), syscall.NLMSG_DONE, 0, uint32(12345), uint32(0), data}
	bs := msg.Bytes()

	// length
	assert(t, bs[0] == 16)

	// type, flags
	assert(t, bs[4] == 3)

	// seq, pid
	assert(t, bs[8] == 57)
	assert(t, bs[9] == 48)
}

func assert(t *testing.T, assertion bool) {
	if !assertion {
		t.Fatalf("assertion failed")
	}
}
