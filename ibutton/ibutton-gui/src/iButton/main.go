package main

import (
	"github.com/lambdasoup/go-netlink/ibutton"
	"gopkg.in/qml.v1"
	"fmt"
)

type App struct {
	ibutton *ibutton.Button
	State   string
	Samples *Samples
}

type Status struct {
	App  *App
	Time string
	Count uint32
	MissionProgress string
	Rate string
	Resolution string
	StartedTime string
}

type Samples struct {
	Len int
	Times []string
	Temps []string
}

func main() {
	err := qml.Run(run)
	if err != nil {
		panic(err)
	}
}

func run() error {
	qml.RegisterTypes("GoExtensions", 1, 0, []qml.TypeSpec{{
		Init: func(s *Status, obj qml.Object) {},
	}, {
		Init: func(a *App, obj qml.Object) {
			a.ibutton = new(ibutton.Button)
			a.State = "DISCONNECTED"
			a.Samples = new(Samples)
		},
	}})

	engine := qml.NewEngine()
	component, err := engine.LoadFile("share/iButton/Main.qml")
	if err != nil {
		panic(err)
	}

	window := component.CreateWindow(nil)
	window.Show()
	window.Wait()

	return nil
}

// Connect the iButton
func (app *App) Connect() {
	go func() {
		app.state("CONNECTING")
		err := app.ibutton.Open()
		if err != nil {
			app.Error()
			app.Disconnect()
			return
		}
		app.state("CONNECTED")
	}()
}

// Disconnect the iButton
func (app *App) Disconnect() {
	app.ibutton.Close()
	app.state("DISCONNECTED")
}

// Start mission
func (app *App) Start() {
	go func() {
		err := app.ibutton.WriteScratchpad()
		if err != nil {
			app.Error()
		}
		data, err := app.ibutton.ReadScratchpad()
		if err != nil {
			app.Error()
		}
		// verify transfer status register
		if data[2] != byte(0x1F) {
			app.Error()
		}
		err = app.ibutton.CopyScratchpad()
		if err != nil {
			app.Error()
		}
		err = app.ibutton.StartMission()
		if err != nil {
			app.Error()
		}
	}()
}

// Stop mission
func (app *App) Stop() {
	go func() {
		app.ibutton.StopMission()
	}()
}

// Clear mission log
func (app *App) Clear() {
	go func() {
		app.ibutton.ClearMemory()
	}()
}

// Error displays an error message
func (app *App) Error() {
	panic("generic error")
}

// Update the button status
func (s *Status) Update() {
	status, err := s.App.ibutton.Status()
	if err != nil {
		s.App.Error()
	}

	s.Time = status.Time().String()
	qml.Changed(s, &s.Time)

	s.Rate = fmt.Sprintf("%v", status.SampleRate())
	qml.Changed(s, &s.Rate)

	if status.HighResolution() {
			s.Resolution = "0.0625°C"
	} else {
			s.Resolution = "0.5°C"
	}
	qml.Changed(s, &s.Resolution)

	s.Count = status.SampleCount()
	qml.Changed(s, &s.Count)

	if status.MissionInProgress() {
		s.MissionProgress = "RUNNING"
	} else {
		s.MissionProgress = "STOPPED"
	}
	qml.Changed(s, &s.MissionProgress)

	s.StartedTime = status.MissionTimestamp().String()
	qml.Changed(s, &s.StartedTime)
}

func (app *App) ReadLog() {
	samples, err := app.ibutton.ReadLog()
	app.Samples.Len = len(samples)
	for _, sample := range samples {
		app.Samples.Times = append(app.Samples.Times, fmt.Sprintf("%v", sample.Time))
		app.Samples.Temps = append(app.Samples.Temps, fmt.Sprintf("%3.3f°C", sample.Temp))
	}
	if err != nil {
		app.Error()
	}
	qml.Changed(app, &app.Samples)
}

func (app *App) SampleTemp(i int) string {
	return fmt.Sprintf("%s", app.Samples.Temps[i])
}

func (app *App) SampleTime(i int) string {
	return fmt.Sprintf("%s", app.Samples.Times[i])
}


func (app *App) state(newState string) {
	app.State = newState
	qml.Changed(app, &app.State)
}
