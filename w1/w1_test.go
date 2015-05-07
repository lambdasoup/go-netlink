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
	"testing"
)

func TestListMasters(t *testing.T) {
	w1 := new(W1)

	err := w1.Open()
	if err != nil {
		t.Fatalf("could not open w1 bus: %v\n", err)
	}
	defer w1.Close()

	response, err := w1.ListMasters()
	if err != nil {
		t.Fatalf("could not request list masters: %v\n", err)
	}

	t.Logf("response: %v\n", response)
}

func TestListSlaves(t *testing.T) {
	w1 := new(W1)

	err := w1.Open()
	if err != nil {
		t.Fatalf("could not open w1 bus: %v\n", err)
	}
	defer w1.Close()

	masters, err := w1.ListMasters()
	if err != nil {
		t.Fatalf("could not request list masters: %v\n", err)
	}
	if len(masters) < 1 {
		t.Fatal("no list masters found")
	}

	response, err := masters[0].ListSlaves()
	if err != nil {
		t.Fatalf("could not request slaves: %v\n", err)
	}

	t.Logf("response: % x\n", response)
}

func TestReadSlave(t *testing.T) {
	w1 := new(W1)

	err := w1.Open()
	if err != nil {
		t.Fatalf("could not open w1 bus: %v\n", err)
	}
	defer w1.Close()

	masters, err := w1.ListMasters()
	if err != nil {
		t.Fatalf("could not request list masters: %v\n", err)
	}
	if len(masters) < 1 {
		t.Fatal("no list masters found")
	}

	slaves, err := masters[0].ListSlaves()
	if err != nil {
		t.Fatalf("could not request slaves: %v\n", err)
	}

	data := make([]byte, 11)
	data[0] = 0x69
	data[1] = 0x00
	data[2] = 0x02

	_, err = slaves[1].Read(data, 108)
	if err != nil {
		t.Fatalf("could not read data: %v\n", err)
	}

}
