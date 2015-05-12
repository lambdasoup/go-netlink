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

// Package main provides a client for Maxim iButton devices
package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/lambdasoup/go-netlink/ibutton"
	"github.com/lambdasoup/go-netlink/log"
)

// parse arguments
var command = flag.String("command", "help", "displays general help")
var logging = flag.Bool("debug", false, "toggle debug logging")

func main() {

	flag.Parse()

	log.SetLogging(*logging)

	switch *command {
	case "status":

		button := new(ibutton.Button)

		if err := button.Open(); err != nil {
			fmt.Printf("could not open iButton (%v)\n", err)
			os.Exit(1)
		}
		defer button.Close()

		status, err := button.Status()
		if err != nil {
			fmt.Printf("could not get iButton status (%v)\n", err)
			os.Exit(1)
		}
		fmt.Printf("time:           %v\n", status.Time())
		fmt.Printf("model:          %v\n", status.Name())
		fmt.Printf("timestamp:      %v\n", status.MissionTimestamp())
		fmt.Printf("count:          %v\n", status.SampleCount())
		fmt.Printf("running:        %v\n", status.MissionInProgress())
		fmt.Printf("memory cleared: %v\n", status.MemoryCleared())
		fmt.Printf("resolution:     %v\n", func() string {
			if status.HighResolution() {
				return "0.0625°C"
			}
			return "0.5°C"
		}())
		fmt.Printf("rate:           %v\n", status.SampleRate())
	case "clear":
		button := new(ibutton.Button)
		err := button.Open()
		defer button.Close()
		if err != nil {
			fmt.Printf("could not open button (%v)\n", err)
			os.Exit(1)
		}
		err = button.ClearMemory()
		if err != nil {
			fmt.Printf("could not clear memory (%v)\n", err)
			os.Exit(1)
		}
		fmt.Printf("Cleared Memory.\n")
	case "start":
		button := new(ibutton.Button)
		err := button.Open()
		defer button.Close()
		if err != nil {
			fmt.Printf("could not open button (%v)\n", err)
			os.Exit(1)
		}
		err = button.WriteScratchpad()
		if err != nil {
			fmt.Printf("could not write scratchpad (%v)\n", err)
			os.Exit(1)
		}
		data, err := button.ReadScratchpad()
		if err != nil {
			fmt.Printf("could not read scratchpad (%v)\n", err)

		}
		// verify transfer status register
		if data[2] != byte(0x1F) {
			fmt.Printf("scratchpad verification failed (%v)\n", data)
			os.Exit(1)
		}
		err = button.CopyScratchpad()
		if err != nil {
			fmt.Printf("could not copy scratchpad (%v)\n", err)
			os.Exit(1)
		}
		err = button.StartMission()
		if err != nil {
			fmt.Printf("could not start mission (%v)\n", err)
			os.Exit(1)
		}
		fmt.Printf("Started mission.\n")
	case "read":
		button := new(ibutton.Button)
		err := button.Open()
		defer button.Close()
		if err != nil {
			fmt.Printf("could not open button (%v)\n", err)
			os.Exit(1)
		}
		samples, err := button.ReadLog()
		if err != nil {
			fmt.Printf("could not read log (%v)\n", err)
			os.Exit(1)
		}
		for _, sample := range samples {
			fmt.Printf("%v\t%3.3f°C\n", sample.Time, sample.Temp)
		}
	case "stop":
		button := new(ibutton.Button)
		err := button.Open()
		defer button.Close()
		if err != nil {
			fmt.Printf("could not open button (%v)\n", err)
			os.Exit(1)
		}
		err = button.StopMission()
		if err != nil {
			fmt.Printf("could not stop mission (%v)\n", err)
			os.Exit(1)
		}
		fmt.Printf("Stopped mission.\n")
	case "help":
		flag.Usage()
		os.Exit(2)
	}

}
