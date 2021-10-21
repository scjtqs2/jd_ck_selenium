package app

import (
	"context"
	"errors"
	"fmt"
	"github.com/guonaihong/gout"
	log "github.com/sirupsen/logrus"
	"github.com/tebeka/selenium"
	"go.uber.org/dig"
	"jd_ck_selenium/util"
	"runtime"
	"time"
)

var chromeVersion = "95.0.4638.17"

var chromeMirrors = "https://npm.taobao.org/mirrors/chromedriver"

type ChromeDriver struct {
	Wd         selenium.WebDriver
	Service    *selenium.Service
	Ct         *dig.Container
	DriverPath string
}

func (ch *ChromeDriver) GetWd() selenium.WebDriver {
	return ch.Wd
}

func (ch *ChromeDriver) GetService() *selenium.Service {
	return ch.Service
}

func (ch *ChromeDriver) GetFileDriverPath() string {
	return ch.DriverPath
}

func NewChromeService(ct *dig.Container) SeInterface {
	return &ChromeDriver{}
}

// 获取cookie并校验cookie 是否存在
func (ch *ChromeDriver) GetCookies(ct *dig.Container) {
	for {
		select {
		case <-c:
			return
		default:
			cks, err := ch.Wd.GetCookies()
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
func (ch *ChromeDriver) GetDriverPath(ct *dig.Container) (string, error) {
	src := ""
	osname := ""
	filename := ""
	bfile := "chromedriver"
	var err error
	chromeVersion, err = ch.CheckLastVersion()
	switch runtime.GOOS {
	case "windows":
		osname = "win32"
		src = fmt.Sprintf("%s/%s/chromedriver_%s.zip", chromeMirrors, chromeVersion, osname)
		filename = fmt.Sprintf("chromedriver_%s.zip", osname)
		bfile = "chromedriver.exe"
		break
	case "darwin":
		osname = "mac64"
		if runtime.GOARCH == "arm64" {
			src = fmt.Sprintf("%s/%s/chromedriver_%s-m1.zip", chromeMirrors, chromeVersion, osname)
			filename = fmt.Sprintf("chromedriver_%s-m1.zip", osname)
		} else {
			src = fmt.Sprintf("%s/%s/chromedriver_%s.zip", chromeMirrors, chromeVersion, osname)
			filename = fmt.Sprintf("chromedriver_%s.zip", osname)
		}
		break
	case "linux":
		osname = "linux64"
		src = fmt.Sprintf("%s/%s/chromedriver_%s.zip", chromeMirrors, chromeVersion, osname)
		filename = fmt.Sprintf("chromedriver_%s.zip", osname)
		break
	default:
		log.Errorf("os =%s,arch=%s \n", runtime.GOOS, runtime.GOARCH)
		return "", errors.New("not support os")
	}
	switch runtime.GOARCH {
	case "arm64":
		if osname != "mac64" {
			return "", errors.New("not support arch")
		}
		break
	case "amd64":
		break
	case "386":
		if osname != "win32" {
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

func (ch *ChromeDriver) SeRun(ct *dig.Container) (err error) {
	p, _ := pickUnusedPort()
	//p := 18777
	opts := []selenium.ServiceOption{
		//selenium.StartFrameBuffer(),           // Start an X frame buffer for the browser to run in.
		//selenium.Output(os.Stderr), // Output debug information to STDERR.
	}
	ch.DriverPath, err = ch.GetDriverPath(ct)
	if err != nil {
		return err
	}
	selenium.SetDebug(false)
	ch.Service, err = selenium.NewChromeDriverService(ch.DriverPath, p, opts...)
	if err != nil {
		return err
	}

	// Connect to the WebDriver instance running locally.
	caps := selenium.Capabilities{"browserName": "chrome"}
	ch.Wd, err = selenium.NewRemote(caps, fmt.Sprintf("http://localhost:%d/wd/hub", p))
	if err != nil {
		return err
	}

	// Navigate to the simple playground interface.
	if err = ch.Wd.Get("https://home.m.jd.com/myJd/newhome.action"); err != nil {
		return err
	}
	go ch.GetCookies(ct)
	return err
}

func (ch *ChromeDriver) CheckLastVersion() (version string, err error) {
	url := "https://npm.taobao.org/mirrors/chromedriver/LATEST_RELEASE"
	code := 0
	err = gout.GET(url).BindBody(&version).Code(&code).
		SetTimeout(timeout).
		F().Retry().Attempt(5).
		WaitTime(time.Millisecond * 500).MaxWaitTime(time.Second * 5).
		Do()
	if err != nil || code != 200 {
		return chromeVersion, err
	}
	log.Infof("latest chrome version =%s ", version)
	return version, err
}
