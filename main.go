package main

import (
	"embed"
	"go.uber.org/dig"
	"jd_ck_selenium/app"
)

// 使用 go 1.16的新特性，自带的打包静态资源的包。
//go:embed static/*
var f embed.FS

var WebHookUrl = ""

var WebHookMethod = "POST"

var WebHookKey = "hhkb"

var SeType = "firefox"

func main() {
	webhook := app.WebHook{
		Url:    WebHookUrl,
		Method: WebHookMethod,
		Key:    WebHookKey,
	}
	container := dig.New()
	container.Provide(func() (static embed.FS) {
		return f
	})
	container.Provide(func() app.WebHook {
		return webhook
	})
	app.SeType=SeType
	app.Run(container)
}
