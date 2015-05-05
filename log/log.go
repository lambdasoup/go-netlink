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

// Package log is a simple wrapper around the standard log package which enables
// logging to be [de]activated
package log

import "log"

var logging = false

// SetLogging [de]activates logging
func SetLogging(newLogging bool) {
	logging = newLogging
}

// Printf log line. Same API as build-in log
func Printf(format string, v ...interface{}) {
	if !logging {
		return
	}

	log.Printf(format, v)
}

// Print log line. Same API as build-in log
func Print(v ...interface{}) {
	if !logging {
		return
	}

	log.Print(v)
}
