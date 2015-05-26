package main

import (
        "gopkg.in/qml.v1"
        "github.com/lambdasoup/go-netlink/ibutton"
)

type App struct {
    *ibutton.Button
    *qml.Context
    State string
}

func main() {
        err := qml.Run(run)
        if err != nil {
                panic(err)
        }
}

func run() error {
        engine := qml.NewEngine()
        component, err := engine.LoadFile("share/iButton/Main.qml")
        if err != nil {
                panic(err)
        }

       app := App{new(ibutton.Button), engine.Context(), "DISCONNECTED"}
       app.SetVar("app", &app)

       window := component.CreateWindow(nil)
       window.Show()
       window.Wait()

       return nil
}

// Connect the iButton
func (app *App) Connect() {
    go func() {
        app.state("CONNECTING")
        err := app.Open()
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
        app.Close()
        app.state("DISCONNECTED")
    }()
}

// Error displays an error message
func (app *App) Error() {
    // TODO show error
}

// Update the button status
func (app *App) Update() {
    go func() {
        status, err := app.Status()
        if err != nil {
            app.Error()
            return
        }
        app.SetVar("status", status.Time().String())
    }()
}

func (app *App) state(newState string) {
    app.State = newState
    qml.Changed(app, &app.State)
}
