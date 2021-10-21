package app

import (
	"github.com/tebeka/selenium"
	"go.uber.org/dig"
)

type SeInterface interface {
	GetCookies(ct *dig.Container)
	GetDriverPath(ct *dig.Container) (string, error)
	SeRun(ct *dig.Container) (error)
	GetWd() selenium.WebDriver
	GetService() *selenium.Service
	GetFileDriverPath() string
}

var SeType string

func NewSeService(ct *dig.Container) (SeInterface, error) {
	switch SeType {
	case "firefox":
		return NewGeckoService(ct), nil
	case "chrome":
		return NewChromeService(ct), nil
	default:
		return NewChromeService(ct), nil
	}
}
