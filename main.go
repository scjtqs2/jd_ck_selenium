package main

import (
	"embed"
	"errors"
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/tebeka/selenium"
	"go.uber.org/dig"
	"io"
	"jd_ck_selenium/util"
	"net"
	"os"
	"os/signal"
	"runtime"
	"time"
)

// 使用 go 1.16的新特性，自带的打包静态资源的包。
//go:embed static/*
var f embed.FS

var WebHookUrl = ""
var c = make(chan os.Signal, 1)

func main() {
	container := dig.New()
	container.Provide(func() embed.FS {
		return f
	})
	port, _ := pickUnusedPort()
	opts := []selenium.ServiceOption{
		//selenium.StartFrameBuffer(),           // Start an X frame buffer for the browser to run in.
		//selenium.Output(os.Stderr), // Output debug information to STDERR.
	}
	geckoDriverPath, err := getGeckoDriverPath(container)
	if err != nil {
		panic(err)
	}
	defer os.Remove(geckoDriverPath)
	selenium.SetDebug(false)
	service, err := selenium.NewGeckoDriverService(geckoDriverPath, port, opts...)
	if err != nil {
		panic(err) // panic is used only as an example and is not otherwise recommended.
	}
	defer service.Stop()

	// Connect to the WebDriver instance running locally.
	caps := selenium.Capabilities{"browserName": "firefox"}
	wd, err := selenium.NewRemote(caps, fmt.Sprintf("http://localhost:%d", port))
	if err != nil {
		panic(err)
	}
	defer wd.Quit()

	// Navigate to the simple playground interface.
	if err := wd.Get("https://home.m.jd.com/myJd/newhome.action"); err != nil {
		panic(err)
	}

	go getCookies(wd, service)

	signal.Notify(c, os.Interrupt, os.Kill)
	<-c
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

// 获取cookie并校验cookie 是否存在
func getCookies(wd selenium.WebDriver, service *selenium.Service) {
	for {
		cks, err := wd.GetCookies()
		var pt_pin, pt_key string
		if err != nil {
			c <- os.Kill
			return
		}
		for _, v := range cks {
			//fmt.Printf("cookie  value=%v", v)
			if v.Name == "pt_pin" {
				pt_pin = v.Value
			}
			if v.Name == "pt_key" {
				pt_key = v.Value
			}
		}
		if pt_pin != "" && pt_key != "" {
			log.Info("############  登录成功，获取到 Cookie  #############")
			log.Infof("cookie=pt_pin=%s, pt_key=%s", pt_pin, pt_key)
			log.Info("####################################################")
			//wd.Quit()
			//service.Stop()
			c <- os.Kill
			return
		}
		time.Sleep(time.Second * 1)
	}
}

// 获取 系统和架构，读取geckodriver的位置

func getGeckoDriverPath(ct *dig.Container) (string, error) {
	path := "static/geckodriver-"
	osname := ""
	arch := ""
	var err error
	var f embed.FS
	ct.Invoke(func(file embed.FS) {
		f = file
	})
	switch runtime.GOOS {
	case "windows":
		osname = "win"
		break
	case "darwin":
		osname = "macos"
		break
	case "linux":
		osname = "linux"
		break
	default:
		log.Errorf("os =%s,arch=%s \n", runtime.GOOS, runtime.GOARCH)
		return "", errors.New("not support os")
	}
	switch runtime.GOARCH {
	case "arm64":
		arch = "arm64"
		if osname != "macos" {
			return "", errors.New("not support arch")
		}
		break
	case "amd64":
		arch = "amd64"
		if osname == "windows" {
			arch = "amd64.exe"
		}
		break
	case "386":
		arch = "i386.exe"
		if osname != "windows" {
			return "", errors.New("not support arch")
		}
		break
	default:
		return "", errors.New("not support arch")
	}
	filepath := path + osname + "-" + arch
	testFile, err := f.Open(filepath)
	if err != nil {
		return "", err
	}
	//将文件拷贝到tmp文件夹下
	defer testFile.Close()
	dst := "./geckodriver-" + osname + "-" + arch
	if !util.PathExists(dst) {
		destination, err := os.Create(dst)
		if err != nil {
			return "", err
		}
		defer destination.Close()
		_, err = io.Copy(destination, testFile)
		destination.Chmod(0755)
	}
	return dst, err
}
