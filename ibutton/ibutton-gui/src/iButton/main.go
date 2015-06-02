package main

import (
	"fmt"
	"github.com/lambdasoup/go-netlink/ibutton"
	"gopkg.in/qml.v1"
)

type App struct {
	ibutton *ibutton.Button
	Connected bool
	Samples *Samples
	Status  *Status
}

type Status struct {
	Time            string
	Count           uint32
	MissionProgress bool
	Rate            string
	Resolution      string
	StartedTime     string
	Cleared         bool
}

type Samples struct {
	Len   int
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
	err := app.ibutton.Open()
	if err != nil {
		app.Error()
		app.Disconnect()
		return
	}
	app.Connected = true
	qml.Changed(app, &app.Connected)
}

// Disconnect the iButton
func (app *App) Disconnect() {
	app.ibutton.Close()
	app.Connected = false
	qml.Changed(app, &app.Connected)
}

// Start mission
func (app *App) Start() {
	err := app.ibutton.ClearMemory()
	if err != nil {
		app.Error()
	}
	err = app.ibutton.WriteScratchpad()
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
	app.Update()
}

// Stop mission
func (app *App) Stop() {
	app.ibutton.StopMission()
	app.Update()
}

// Error displays an error message
func (app *App) Error() {
	panic("generic error")
}

// Update the button status
func (app *App) Update() {
	status, err := app.ibutton.Status()
	if err != nil {
		app.Error()
	}

	app.Status.Time = status.Time().String()
	qml.Changed(app.Status, &app.Status.Time)

	app.Status.Rate = fmt.Sprintf("%v", status.SampleRate())
	qml.Changed(app.Status, &app.Status.Rate)

	if status.HighResolution() {
		app.Status.Resolution = "0.0625°C"
	} else {
		app.Status.Resolution = "0.5°C"
	}
	qml.Changed(app.Status, &app.Status.Resolution)

	app.Status.Count = status.SampleCount()
	qml.Changed(app.Status, &app.Status.Count)

	app.Status.MissionProgress = status.MissionInProgress()
	qml.Changed(app.Status, &app.Status.MissionProgress)

	app.Status.StartedTime = status.MissionTimestamp().String()
	qml.Changed(app.Status, &app.Status.StartedTime)
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
