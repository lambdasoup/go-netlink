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
	"os/exec"
	"testing"
)

const (
	CN_NETLINK_USERS = 11 // uapi/conn.h
	CN_TEST_IDX      = CN_NETLINK_USERS + 3
	CN_TEST_VAL      = 0x456
)

var moduleLoaded = false

func loadModule(t *testing.T) {
	if moduleLoaded {
		return
	}

	cmd := exec.Command("sudo", "insmod", "module/cn_test.ko")
	err := cmd.Start()
	if err != nil {
		t.Logf("Could not load kernel module: %v", err)
	}
	t.Log("Loading kernel module")
	err = cmd.Wait()
	if err != nil {
		t.Logf("Loading finished with error: %v", err)
	}
	moduleLoaded = true
}

func unloadModule(t *testing.T) {
	cmd := exec.Command("sudo", "rmmod", "cn_test")
	err := cmd.Start()
	if err != nil {
		t.Logf("Could not unload kernel module: %v", err)
	}
	t.Log("Unloading kernel module")
	err = cmd.Wait()
	if err != nil {
		t.Logf("Unloading finished with error: %v", err)
	}
	moduleLoaded = false
}

func TestRequestResponse(t *testing.T) {
	loadModule(t)
	defer unloadModule(t)

	c, err := Open(CbId{CN_TEST_IDX, CN_TEST_VAL})
	if err != nil {
		t.Fatalf("could not open the connector: %v", err)
	}
	defer c.Close()

	data := new(bytes.Buffer)
	data.WriteString("test data2")

	response, err := c.Request(data.Bytes())
	if err != nil {
		t.Fatalf("could not complete request: %v", err)
	}

	t.Logf("response data: %v", response)
}

func TestParseConnectorMessage(t *testing.T) {
	bs := make([]byte, 0)

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

	assert(t, msg.id.Idx == CN_TEST_IDX)
	assert(t, msg.id.Val == CN_TEST_VAL)

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
