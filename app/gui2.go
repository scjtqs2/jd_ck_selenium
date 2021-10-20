package app

import (
	"fmt"
	"github.com/asticode/go-astikit"
	"github.com/asticode/go-astilectron"
	log "github.com/sirupsen/logrus"
	"go.uber.org/dig"
	"os"
)

func guiStart2(port int, ct *dig.Container)  {
	// Initialize astilectron
	var a, _ = astilectron.New(log.New(), astilectron.Options{
		AppName:           "jd_cookie_Tools",
		BaseDirectoryPath: "example",
	})
	defer a.Close()
	var err error
	// Handle signals
	//a.HandleSignals()
	// Start astilectron
	// Start
	if err = a.Start(); err != nil {
		log.Fatalf("main: starting astilectron failed: %w", err)
	}

	// New window
	var w *astilectron.Window
	if w, err = a.NewWindow(fmt.Sprintf("http://127.0.0.1:%d/", port), &astilectron.WindowOptions{
		Center: astikit.BoolPtr(true),
		Height: astikit.IntPtr(600),
		Width:  astikit.IntPtr(800),
	}); err != nil {
		log.Fatal(fmt.Errorf("main: new window failed: %w", err))
	}
	// This will listen to messages sent by Javascript
	w.OnMessage(func(m *astilectron.EventMessage) interface{} {
		// Unmarshal
		var s string
		m.Unmarshal(&s)
		//log.Info(s)
		// Process message
		switch s {
		case "quit":
			os.Remove(geckoDriverPath)
			service.Stop()
			wd.Quit()
			c <- os.Kill
			w.Destroy()
			a.Close()
			break
		case "open":
			seRun(ct)
			break
		case "getck":
			cookie, err := cache.Get(cache_key_cookie)
			if err != nil {
				return ""
			}
			return cookie.(string)
		}
		return "success"
	})
	// Open dev tools
	// Create windows
	if err := w.Create(); err != nil {
		log.Fatal(fmt.Errorf("main: creating window failed: %w", err))
	}
	w.OpenDevTools()



	// Blocking pattern
	a.Wait()
}