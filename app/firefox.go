package app

import (
	"context"
	"errors"
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/tebeka/selenium"
	"go.uber.org/dig"
	"jd_ck_selenium/util"
	"runtime"
	"time"
)

var geckoVersion = "v0.30.0"

var geckoMirrors = "https://npm.taobao.org/mirrors/geckodriver"

type GeckoDriver struct {
	Wd         selenium.WebDriver
	Service    *selenium.Service
	Ct         *dig.Container
	DriverPath string
}

var geckoDriver = &GeckoDriver{}

// 获取cookie并校验cookie 是否存在
func (ge *GeckoDriver) GetCookies(ct *dig.Container) {
	for {
		select {
		case <-c:
			return
		default:
			cks, err := ge.Wd.GetCookies()
			var pt_pin, pt_key string
			if err != nil {
				return
			}
			for _, v := range cks {
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
				cookie := fmt.Sprintf("pt_pin=%s, pt_key=%s", pt_pin, pt_key)
				cache.Set(cache_key_cookie, cookie)
				postWebHookCk(ct, cookie)
				return
			}
		}
		time.Sleep(time.Microsecond * 300)
	}
}

// 获取 系统和架构，读取geckodriver的位置
func (ge *GeckoDriver) GetGeckoDriverPath(ct *dig.Container) (string, error) {
	src := ""
	osname := ""
	filename := ""
	bfile := "geckodriver"
	var err error
	switch runtime.GOOS {
	case "windows":
		osname = "win"
		switch runtime.GOOS {
		case "amd64":
			src = fmt.Sprintf("%s/%s/geckodriver-%s-%s64.tar.gz", geckoMirrors, geckoVersion, geckoVersion, osname)
			filename = fmt.Sprintf("geckodriver-%s-%s64.tar.gz", geckoVersion, osname)
			break
		case "386":
			src = fmt.Sprintf("%s/%s/geckodriver-%s-%s32.tar.gz", geckoMirrors, geckoVersion, geckoVersion, osname)
			filename = fmt.Sprintf("geckodriver-%s-%s32.tar.gz", geckoVersion, osname)
			break
		default:
			return "", errors.New("not support os")
		}
		bfile = "geckodriver.exe"
		break
	case "darwin":
		osname = "macos"
		if runtime.GOOS == "arm64" {
			src = fmt.Sprintf("%s/%s/geckodriver-%s-%s-aarch64.tar.gz", geckoMirrors, geckoVersion, geckoVersion, osname)
			filename = fmt.Sprintf("geckodriver-%s-%s-aarch64.tar.gz", geckoVersion, osname)
		} else {
			src = fmt.Sprintf("%s/%s/geckodriver-%s-%s.tar.gz", geckoMirrors, geckoVersion, geckoVersion, osname)
			filename = fmt.Sprintf("geckodriver-%s-%s.tar.gz", geckoVersion, osname)
		}
		break
	case "linux":
		osname = "linux"
		switch runtime.GOOS {
		case "amd64":
			src = fmt.Sprintf("%s/%s/geckodriver-%s-%s64.tar.gz", geckoMirrors, geckoVersion, geckoVersion, osname)
			filename = fmt.Sprintf("geckodriver-%s-%s64.tar.gz", geckoVersion, osname)
			break
		case "386":
			src = fmt.Sprintf("%s/%s/geckodriver-%s-%s32.tar.gz", geckoMirrors, geckoVersion, geckoVersion, osname)
			filename = fmt.Sprintf("geckodriver-%s-%s32.tar.gz", geckoVersion, osname)
			break
		default:
			return "", errors.New("not support os")
		}
		break
	default:
		log.Errorf("os =%s,arch=%s \n", runtime.GOOS, runtime.GOARCH)
		return "", errors.New("not support os")
	}
	switch runtime.GOARCH {
	case "arm64":
		if osname != "macos" {
			return "", errors.New("not support arch")
		}
		break
	case "amd64":
		break
	case "386":
		if osname != "win" {
			return "", errors.New("not support arch")
		}
		break
	default:
		return "", errors.New("not support arch")
	}
	dst := "./tmp"
	util.Download(context.Background(), src, fmt.Sprintf("%s/%s", dst, filename))
	util.Unpack(context.Background(), fmt.Sprintf("%s/%s", dst, filename), dst)
	return fmt.Sprintf("%s/%s", dst, bfile), err
}

func (ge *GeckoDriver) seRun(ct *dig.Container) {
	p, _ := pickUnusedPort()
	//p := 18777
	opts := []selenium.ServiceOption{
		//selenium.StartFrameBuffer(),           // Start an X frame buffer for the browser to run in.
		//selenium.Output(os.Stderr), // Output debug information to STDERR.
	}
	var err error
	//defer func() {
	//	c <- os.Kill
	//}()
	ge.DriverPath, err = ge.GetGeckoDriverPath(ct)
	if err != nil {
		panic(err)
	}
	selenium.SetDebug(false)
	ge.Service, err = selenium.NewGeckoDriverService(ge.DriverPath, p, opts...)
	if err != nil {
		panic(err) // panic is used only as an example and is not otherwise recommended.
	}

	// Connect to the WebDriver instance running locally.
	caps := selenium.Capabilities{"browserName": "firefox"}
	ge.Wd, err = selenium.NewRemote(caps, fmt.Sprintf("http://localhost:%d", p))
	if err != nil {
		panic(err)
	}

	// Navigate to the simple playground interface.
	if err := ge.Wd.Get("https://home.m.jd.com/myJd/newhome.action"); err != nil {
		panic(err)
	}
	go ge.GetCookies(ct)
}
