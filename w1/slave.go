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

import "fmt"

// Slave is a 1-Wire slave device
type Slave struct {
	family byte
	uid    [6]byte
	crc    byte
	master *Master
}

// Close this Slave's 1-Wire connection
func (s *Slave) Close() {
	s.master.Close()
}

// IsFamily returns true if this Slave is of the given device family
func (s *Slave) IsFamily(family byte) bool {
	return s.family == family
}

func (s *Slave) String() string {
	return fmt.Sprintf("Slave{Family:%x, UID:%x}", s.family, s.uid)
}

func (s *Slave) Read(data []byte, count int) ([]byte, error) {
	return s.master.readSlave(s, data, count)
}

func (s *Slave) Write(data []byte) error {
	return s.master.writeSlave(s, data)
}
