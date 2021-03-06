package app

import (
	"go.uber.org/dig"
	"net"
	"os"
	"os/signal"
)

var c = make(chan os.Signal, 1)

var svc SeInterface

func Run(ct *dig.Container) {
	svc, _ = NewSeService(ct)
	//启动gin的http服务
	httpPort := HTTPServer.httpStart(ct)
	go func() {
		signal.Notify(c, os.Interrupt, os.Kill)
		<-c
		svc.GetWd().Quit()
		svc.GetService().Stop()
		os.Remove(svc.GetFileDriverPath())
	}()
	guiStart(httpPort, ct)

}

func pickUnusedPort() (int, error) {
	addr, err := net.ResolveTCPAddr("tcp", "127.0.0.1:0")
	if err != nil {
		return 0, err
	}

	l, err := net.ListenTCP("tcp", addr)
	if err != nil {
		return 0, err
	}
	port := l.Addr().(*net.TCPAddr).Port
	if err := l.Close(); err != nil {
		return 0, err
	}
	return port, nil
}
