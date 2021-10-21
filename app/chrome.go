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

var chromeVersion = "95.0.4638.17"

var chromeMirrors = "https://npm.taobao.org/mirrors/chromedriver"

type ChromeDriver struct {
	Wd         selenium.WebDriver
	Service    *selenium.Service
	Ct         *dig.Container
	DriverPath string
}

var chromeDriver = &ChromeDriver{}

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
func (ch *ChromeDriver) GetChromeDriverPath(ct *dig.Container) (string, error) {
	src := ""
	osname := ""
	filename := ""
	bfile := "chromedriver"
	var err error
	switch runtime.GOOS {
	case "windows":
		osname = "win32"
		src = fmt.Sprintf("%s/%s/chromedriver_%s.zip", chromeMirrors, chromeVersion, osname)
		filename = fmt.Sprintf("chromedriver_%s.zip", osname)
		bfile = "chromedriver.exe"
		break
	case "darwin":
		osname = "mac64"
		if runtime.GOOS == "arm64" {
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
	util.Unzip(context.Background(), fmt.Sprintf("%s/%s", dst, filename), dst)
	return fmt.Sprintf("%s/%s", dst, bfile), err
}

func (ch *ChromeDriver) seRun(ct *dig.Container) {
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
	ch.DriverPath, err = ch.GetChromeDriverPath(ct)
	if err != nil {
		panic(err)
	}
	selenium.SetDebug(false)
	ch.Service, err = selenium.NewChromeDriverService(ch.DriverPath, p, opts...)
	if err != nil {
		panic(err) // panic is used only as an example and is not otherwise recommended.
	}

	// Connect to the WebDriver instance running locally.
	caps := selenium.Capabilities{"browserName": "chrome"}
	ch.Wd, err = selenium.NewRemote(caps, fmt.Sprintf("http://localhost:%d/wd/hub", p))
	if err != nil {
		panic(err)
	}

	// Navigate to the simple playground interface.
	if err := ch.Wd.Get("https://home.m.jd.com/myJd/newhome.action"); err != nil {
		panic(err)
	}
	go ch.GetCookies(ct)
}
