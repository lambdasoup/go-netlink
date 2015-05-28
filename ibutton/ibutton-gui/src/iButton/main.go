package main

import (
	"github.com/lambdasoup/go-netlink/ibutton"
	"gopkg.in/qml.v1"
)

type App struct {
	ibutton *ibutton.Button
	State   string
}

type Status struct {
	App  *App
	Time string
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
	go func() {
		app.ibutton.Close()
		app.state("DISCONNECTED")
	}()
}

// Error displays an error message
func (app *App) Error() {
	// TODO show error
}

// Update the button status
func (s *Status) Update() {
	status, err := s.App.ibutton.Status()
	if err != nil {
		s.App.Error()
	}

	s.Time = status.Time().String()
	qml.Changed(s, &s.Time)
}

func (app *App) state(newState string) {
	app.State = newState
	qml.Changed(app, &app.State)
}
