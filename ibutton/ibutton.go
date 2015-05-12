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

// Package ibutton provides access to Maxim iButton devices
package ibutton

import (
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/lambdasoup/go-netlink/log"
	"github.com/lambdasoup/go-netlink/w1"
)

// iButton command codes
const (
	WRITE_SCRATCHPAD uint8 = 0x0F
	COPY_SCRATCHPAD  uint8 = 0x99
	READ_SCRATCHPAD  uint8 = 0xAA
	READ_MEMORY      uint8 = 0x69
	CLEAR_MEMORY     uint8 = 0x96
	STOP_MISSION     uint8 = 0x33
	START_MISSION    uint8 = 0xCC
)

// device identifiers type
type deviceId int

// device identifier byte descriptors
const (
	DS2422  deviceId = 0x00
	DS1923  deviceId = 0x20
	DS1922L deviceId = 0x40
	DS1922T deviceId = 0x60
	DS1922E deviceId = 0x80
)

// device specific data
var devices = map[deviceId]struct {
	name      string
	offset    float32
	supported bool
	tr1       Temperature
}{
	DS2422:  {"DS2422", 0.0, false, 0.0},
	DS1923:  {"DS1923", 0.0, false, 0.0},
	DS1922L: {"DS1922L", -41.0, true, 60.0},
	DS1922T: {"DS1922T", -1.0, true, 90.0},
	DS1922E: {"DS1922E", 0.0, false, 0.0},
}

// Button represents an iButton
type Button struct {
	slave *w1.Slave
}

// Sample represents a mission log sample
type Sample struct {
	Time time.Time
	Temp Temperature
}

// Temperature represents a temperature
type Temperature float32

// Status returns the current iButton status
func (b *Button) Status() (status *Status, err error) {
	status = new(Status)

	status.bytes, err = b.readMemory(0x0200, 3)
	if err != nil {
		return
	}

	return
}

// Open opens this iButton's 1-Wire session
func (b *Button) Open() (err error) {

	// open 1-wire connection
	w1 := new(w1.W1)
	err = w1.Open()
	if err != nil {
		return
	}

	// find master
	ms, err := w1.ListMasters()
	if err != nil {
		err = fmt.Errorf("could not request list masters: %v\n", err)
		return
	}
	if len(ms) < 1 {
		err = fmt.Errorf("no list masters found")
		return
	}

	// find ibutton slave
	ss, err := ms[0].ListSlaves()
	if err != nil {
		err = fmt.Errorf("could not request slaves: %v\n", err)
		return
	}
	ss = filterFamily(ss)
	if len(ss) < 1 {
		err = fmt.Errorf("no ibuttons found")
		return
	}
	b.slave = &ss[0]

	return
}

func filterFamily(ss []w1.Slave) (filtered []w1.Slave) {
	for _, s := range ss {
		if s.IsFamily(0x41) {
			filtered = append(filtered, s)
		}
	}
	return
}

func (b *Button) String() string {
	return fmt.Sprintf("Button{%v}", b.slave)
}

// Close closes this iButton's 1-Wire session
func (b *Button) Close() {
	b.slave.Close()
}

// StopMission stops the currently running mission
func (b *Button) StopMission() error {
	data := make([]byte, 10)
	data[0] = STOP_MISSION
	data[9] = 0xFF
	return b.slave.Write(data)
}

// ClearMemory clears the ibutton memory
func (b *Button) ClearMemory() error {
	data := make([]byte, 10)
	data[0] = CLEAR_MEMORY
	data[9] = 0xFF
	return b.slave.Write(data)
}

// StartMission starts a mission
func (b *Button) StartMission() error {
	data := make([]byte, 10)
	data[0] = START_MISSION
	data[9] = 0xFF
	return b.slave.Write(data)
}

// CopyScratchpad copies the scratchpad
func (b *Button) CopyScratchpad() error {
	data := make([]byte, 12)
	data[0] = COPY_SCRATCHPAD
	data[1] = 0x00
	data[2] = 0x02
	data[3] = 0x1F
	return b.slave.Write(data)
}

// WriteScratchpad writes the button scrathpad
func (b *Button) WriteScratchpad() error {
	data := make([]byte, 35)

	// command00 02
	data[0] = WRITE_SCRATCHPAD

	// target address (scratchpad)
	data[1] = 0x00
	data[2] = 0x02

	// time and date (01.04.2013 15:30:00)
	// strange format, so: 30 -> "30" -> 0x30
	now := time.Now()
	second, _ := strconv.ParseInt(strconv.Itoa(now.Second()), 16, 8)
	minute, _ := strconv.ParseInt(strconv.Itoa(now.Minute()), 16, 8)
	hour, _ := strconv.ParseInt(strconv.Itoa(now.Hour()), 16, 8)
	data[3] = byte(second)
	data[4] = byte(minute)
	data[5] = byte(hour)
	data[6] = byte(now.Day())
	data[7] = byte(now.Month())
	data[8] = byte(now.Year() % 100)

	// sample rate (10mins with EHSS=0)
	data[9] = 0x0A
	data[10] = 0x00

	// alarm thresholds
	data[11] = 0x52
	data[12] = 0x99

	// alarm control (both disabled = 0)
	data[19] = 0x00

	// "Disabled" - registers is R/W but should be 0xfc
	data[20] = 0xFC

	// EHSS=0 (low sample rate), EOSC=1 (oscillator running)
	data[21] = 0x01

	// no alarm, no rollover, 16 bit, logging on
	data[22] = 0xC5

	// no mission start delay
	data[25] = 0x00
	data[26] = 0x00
	data[27] = 0x00

	// "write through the end of the scratchpad"
	data[28] = 0xFF
	data[29] = 0xFF
	data[30] = 0xFF
	data[31] = 0xFF
	data[32] = 0xFF
	data[33] = 0xFF
	data[34] = 0xFF

	return b.slave.Write(data)
}

// ReadScratchpad reads the button scrathpad
func (b *Button) ReadScratchpad() (data []byte, err error) {
	// send the read scratchpad command
	cmd := make([]byte, 1)
	cmd[0] = READ_SCRATCHPAD
	return b.slave.Read(cmd, 35)
}

// ReadLog returns the log entries for the current mission
func (b *Button) ReadLog() (samples []Sample, err error) {

	// aquire button status
	status, err := b.Status()
	if err != nil {
		return
	}

	// make array with sample count length
	samples = make([]Sample, status.SampleCount())

	// determine temperature sample size
	var sampleBytes uint32
	if status.HighResolution() {
		sampleBytes = 2
	} else {
		sampleBytes = 1
	}

	// determine page count
	byteCount := status.SampleCount() * sampleBytes
	pages := int(byteCount / 32)
	if byteCount%32 != 0 {
		pages++
	}

	// read pages from device memory
	bytes, err := b.readMemory(0x1000, pages)
	if err != nil {
		return
	}

	// get temperature correction factors
	A, B, C := status.correctionFactors()

	// parse temperatures
	for index := uint32(0); index < status.SampleCount(); index++ {

		samples[index].Time = status.MissionTimestamp().Add(status.SampleRate() * time.Duration(index))

		temperatureBytes := bytes[index*sampleBytes : (index+1)*sampleBytes]

		tc := status.decodeTemp(temperatureBytes)
		samples[index].Temp = tc - (A*tc*tc + B*tc + C)

	}

	return
}

// ReadMemory reads the iButton's memory starting with the given address
func (b *Button) readMemory(address uint16, pages int) (result []byte, err error) {

	// send the read command
	cmd := make([]byte, 11)
	cmd[0] = READ_MEMORY
	cmd[1] = byte(address)
	cmd[2] = byte(address >> 8)

	data, err := b.slave.Read(cmd, pages*34)
	if err != nil {
		return
	}

	result = make([]byte, pages*32)

	// initial block has special crc checking
	block := make([]byte, 3+34)
	copy(block, cmd[:3])
	copy(block[3:], data[:34])
	checksum := 0xffff ^ (uint16(block[33+3])<<8 + uint16(block[32+3]))
	if Checksum(block[:32+3]) != checksum {
		err = errors.New("crc check failed in initial crc")
		return
	}
	copy(result, block[3:32+3])

	log.Printf("page 1 %x\n", block)
	log.Printf("result %x\n", result)

	// read remaining pages
	for page := 2; page <= pages; page++ {
		block = make([]byte, 34)
		copy(block, data[(page-1)*34:])
		log.Printf("page %d %x\n", page, block)
		checksum := 0xffff ^ (uint16(block[33])<<8 + uint16(block[32]))
		if Checksum(block[:32]) != checksum {
			err = errors.New("crc check failed failed in subsequent crc")
			return
		}
		copy(result[(page-1)*32:], block[:32])
		log.Printf("result %x\n", result)
	}

	return
}
