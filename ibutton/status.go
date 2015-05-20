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
	"fmt"
	"time"
)

// Status represents an iButton status. The iButton's status is saved in two register pages (0x0200-0x0263)
type Status struct {
	bytes []byte
}

// Time the time
func (s *Status) Time() time.Time {

	return parseTime(s.bytes[0x00:0x06])

}

// MissionTimestamp the current mission timestamp
func (s *Status) MissionTimestamp() time.Time {

	return parseTime(s.bytes[0x19:0x1F])

}

// decodeTemp gives the temperature encoded in the given byte slice
func (s *Status) decodeTemp(bytes []byte) (temp Temperature) {

	switch len(bytes) {
	case 1:
		temp = Temperature(float32(bytes[0])/2 + devices[s.DeviceID()].offset)
	case 2:
		temp = Temperature(float32(bytes[0])/2 + devices[s.DeviceID()].offset + float32(bytes[1])/512)
	}

	return
}

// SampleCount count of recorded samples since last mission start
func (s *Status) SampleCount() uint32 {

	return uint32(s.bytes[0x22])<<16 + uint32(s.bytes[0x21])<<8 + uint32(s.bytes[0x20])
}

// MissionInProgress true if a mission is running
func (s *Status) MissionInProgress() bool {

	return s.bytes[0x15]&(0x01<<1) > 0
}

// HighResolution true if the chip is in 16bit (0.0625°C) mode
func (s *Status) HighResolution() bool {

	return s.bytes[0x13]&(0x01<<2) > 0
}

// SampleRate return the currently set sample rate
func (s *Status) SampleRate() (duration time.Duration) {

	// first read in the raw rate
	rate := uint32(s.bytes[0x06]) + uint32(s.bytes[0x07])<<8

	// decide on minutes or seconds
	if s.bytes[0x12]>>1 == 1 {
		duration = time.Duration(rate) * time.Second
	} else {
		duration = time.Duration(rate) * time.Minute
	}

	return
}

// MemoryCleared true when the memory has been successfully cleared (MEMCLR==1)
func (s *Status) MemoryCleared() bool {

	return s.bytes[0x15]&(0x01<<3) > 0
}

// DeviceID returns the device identifier
func (s *Status) DeviceID() (model deviceID) {

	return deviceID(s.bytes[0x26])
}

// correctionFactors returns the temperature correction factors for this device
func (s *Status) correctionFactors() (a Temperature, b Temperature, c Temperature) {

	// get chip-hardcoded correction values
	tr1 := devices[s.DeviceID()].tr1
	tr2 := s.decodeTemp(s.bytes[0x40:0x42])
	tc2 := s.decodeTemp(s.bytes[0x42:0x44])
	tr3 := s.decodeTemp(s.bytes[0x44:0x46])
	tc3 := s.decodeTemp(s.bytes[0x46:0x48])

	// calculate correction factors
	err2 := tc2 - tr2
	err3 := tc3 - tr3
	err1 := err2

	// formula stuff from DS1922L data sheet (p.50)
	b = (tr2*tr2 - tr1*tr1) * (err3 - err1) / ((tr2*tr2-tr1*tr1)*(tr3-tr1) + (tr3*tr3-tr1*tr1)*(tr1-tr2))
	a = b * (tr1 - tr2) / (tr2*tr2 - tr1*tr1)
	c = err1 - a*tr1*tr1 - b*tr1

	return
}

// Name the device model's name
func (s *Status) Name() string {

	device, ok := devices[s.DeviceID()]
	if ok {
		return device.name
	}

	return fmt.Sprintf("Unknown Device (deviceID:%x)", s.DeviceID())
}
