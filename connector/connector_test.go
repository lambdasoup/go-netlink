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

package connector

import (
	"bytes"
	"testing"
)

const (
	cnNetlinkUsers = 11
	cnTestIdx      = cnNetlinkUsers + 3
	cnTestVal      = 0x456
)

func TestParseConnectorMessage(t *testing.T) {
	var bs []byte

	// CB_IDX, CB_VAL
	bs = append(bs, 14, 0, 0, 0, 86, 4, 0, 0)
	// seq
	bs = append(bs, 57, 48, 0, 0)
	// ack
	bs = append(bs, 58, 48, 0, 0)
	// len, flags
	bs = append(bs, 10, 0, 0, 0)
	// payload "test reply"
	payload := []byte{116, 101, 115, 116, 32, 114, 101, 112, 108, 121}
	bs = append(bs, payload...)

	msg, err := parseConnectorMsg(bs)
	if err != nil {
		t.Fatalf("could not parse message: %v", err)
	}

	t.Log(msg)

	assert(t, msg.id.idx == cnTestIdx)
	assert(t, msg.id.val == cnTestVal)

	assert(t, msg.seq == uint32(12345))
	assert(t, msg.ack == uint32(12346))
	assert(t, msg.len == uint16(10))
	assert(t, msg.flags == uint16(0))
	assert(t, bytes.Equal(msg.data, payload))
}

func assert(t *testing.T, assertion bool) {
	if !assertion {
		t.Fatalf("assertion failed")
	}
}
