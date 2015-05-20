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

package ibutton

import (
	"testing"
	"time"
)

func TestParseTime(t *testing.T) {
	bs := make([]byte, 6)

	// 41 seconds - 0100 0001
	bs[0] = 0x41

	// day 28
	bs[3] = 0x28

	time := parseTime(bs)

	if time.Second() != 41 {
		t.Fail()
	}

	if time.Day() != 28 {
		t.Fail()
	}
}

func TestSerializeTime(t *testing.T) {
	bs := make([]byte, 6)

	time := time.Date(2015, 11, 15, 16, 53, 31, 0, time.UTC)
	serializeTime(bs, &time)

	if bs[3] != 0x15 {
		t.Fail()
	}

	if bs[4] != 0x11 {
		t.Fail()
	}
}
