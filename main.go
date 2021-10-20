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

func main() {
	container := dig.New()
	container.Provide(func() (static embed.FS) {
		return f
	})
	container.Provide(func() (WebHookUrl string) {
		return WebHookUrl
	})
	app.Copydll(container)
	app.Run(container)
}
