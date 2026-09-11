package configuration

import (
	"fmt"

	"github.com/labstack/echo"
)

type Application struct {
	server *echo.Echo
}

func NewApplication() *Application {
	server := echo.New()
	server.HideBanner = true
	server.HidePort = true

	return &Application{
		server: server,
	}
}

func (a *Application) Start() {
	fmt.Println("explicAI is starting on 0.0.0.0:8080")
	a.server.Start("0.0.0.0:8080")
}
