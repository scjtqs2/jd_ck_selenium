package app

import (
	"fmt"
	"github.com/webview/webview"
	"go.uber.org/dig"
	"os"
)

var gui webview.WebView

func guiStart(port int, ct *dig.Container) {
	debug := true
	gui = webview.New(debug)
	defer func() {
		c <- os.Kill
	}()
	gui.SetTitle("jd壕羊毛第五大队")
	gui.SetSize(800, 600, webview.HintNone)
	gui.Navigate(fmt.Sprintf("http://127.0.0.1:%d/", port))
	gui.Bind("quit", func() {
		os.Remove(geckoDriverPath)
		service.Stop()
		wd.Quit()
		c <- os.Kill
		gui.Destroy()
	})
	gui.Bind("open", func() {
		seRun(ct)
	})
	gui.Bind("getck", func() string {
		cookie, err := cache.Get(cache_key_cookie)
		if err != nil {
			return ""
		}
		return cookie.(string)
	})
	gui.Run()
}
