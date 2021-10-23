package app

import (
	"context"
	"embed"
	"errors"
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/tebeka/selenium"
	"github.com/tebeka/selenium/firefox"
	"go.uber.org/dig"
	"io"
	"jd_ck_selenium/util"
	"os"
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

func (ge *GeckoDriver) GetWd() selenium.WebDriver {
	return ge.Wd
}

func (ge *GeckoDriver) GetService() *selenium.Service {
	return ge.Service
}

func (ge *GeckoDriver) GetFileDriverPath() string {
	return ge.DriverPath
}

func NewGeckoService(ct *dig.Container) SeInterface {
	return &GeckoDriver{}
}

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
				log.Infof("cookie=pt_pin=%s; pt_key=%s", pt_pin, pt_key)
				log.Info("####################################################")
				cookie := fmt.Sprintf("pt_pin=%s;pt_key=%s;", pt_pin, pt_key)
				cache.Set(cache_key_cookie, cookie)
				postWebHookCk(ct, cookie)
				return
			}
		}
		time.Sleep(time.Microsecond * 300)
	}
}

// 获取 系统和架构，读取geckodriver的位置
func (ge *GeckoDriver) GetDriverPath(ct *dig.Container) (string, error) {
	src := ""
	osname := ""
	filename := ""
	bfile := "geckodriver"
	var err error
	switch runtime.GOOS {
	case "windows":
		osname = "win"
		switch runtime.GOARCH {
		case "amd64":
			src = fmt.Sprintf("%s/%s/geckodriver-%s-%s64.zip", geckoMirrors, geckoVersion, geckoVersion, osname)
			filename = fmt.Sprintf("geckodriver-%s-%s64.zip", geckoVersion, osname)
			break
		case "386":
			src = fmt.Sprintf("%s/%s/geckodriver-%s-%s32.zip", geckoMirrors, geckoVersion, geckoVersion, osname)
			filename = fmt.Sprintf("geckodriver-%s-%s32.zip", geckoVersion, osname)
			break
		default:
			return "", errors.New("not support os")
		}
		bfile = "geckodriver.exe"
		break
	case "darwin":
		osname = "macos"
		if runtime.GOARCH == "arm64" {
			src = fmt.Sprintf("%s/%s/geckodriver-%s-%s-aarch64.tar.gz", geckoMirrors, geckoVersion, geckoVersion, osname)
			filename = fmt.Sprintf("geckodriver-%s-%s-aarch64.tar.gz", geckoVersion, osname)
		} else {
			src = fmt.Sprintf("%s/%s/geckodriver-%s-%s.tar.gz", geckoMirrors, geckoVersion, geckoVersion, osname)
			filename = fmt.Sprintf("geckodriver-%s-%s.tar.gz", geckoVersion, osname)
		}
		break
	case "linux":
		osname = "linux"
		switch runtime.GOARCH {
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

func (ge *GeckoDriver) SeRun(ct *dig.Container) (err error) {
	p, _ := pickUnusedPort()
	//p := 18777
	opts := []selenium.ServiceOption{
		//selenium.StartFrameBuffer(),           // Start an X frame buffer for the browser to run in.
		//selenium.Output(os.Stderr), // Output debug information to STDERR.
	}
	//firefoxPath := ge.checkFirefox(ct)
	//log.Infof("firefoxPath= %s", firefoxPath)
	firefoxPath := ""
	ge.DriverPath, err = ge.GetDriverPath(ct)
	if err != nil {
		return err
	}
	selenium.SetDebug(true)
	ge.Service, err = selenium.NewGeckoDriverService(ge.DriverPath, p, opts...)
	if err != nil {
		return err
	}

	// Connect to the WebDriver instance running locally.
	//caps := selenium.Capabilities{"browserName": "firefox"}
	caps := selenium.Capabilities{}
	ge.Wd, err = selenium.NewRemote(caps, fmt.Sprintf("http://localhost:%d", p))
	if err != nil {
		return err
	}

	//firfox参数
	firefoxCaps := firefox.Capabilities{
		Binary: firefoxPath,
		//Args: []string{
		//	"--user-agent=Mozilla/5.0 (iPhone; CPU iPhone OS 14_6 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/14.0.3 Mobile/15E148 Safari/604.1",
		//	"--window-size=375,812",
		//},
	}
	caps.AddFirefox(firefoxCaps)
	//调整浏览器长宽高
	ge.Wd.ResizeWindow("", 375, 812)

	// Navigate to the simple playground interface.
	if err = ge.Wd.Get("https://home.m.jd.com/myJd/newhome.action"); err != nil {
		return err
	}
	go ge.GetCookies(ct)
	return err
}

func (ge *GeckoDriver) CheckLastVersion() (version string, err error) {
	return geckoVersion, nil
}

//解压内置的firefox到tmp下去
func (ge *GeckoDriver) checkFirefox(ct *dig.Container) string {
	log.Info("解压firefox中，请稍等")
	var f embed.FS
	ct.Invoke(func(static embed.FS) {
		f = static
	})
	pwd := util.GetPwdPath()
	file := "firefox.zip"
	testFile, _ := f.Open(fmt.Sprintf("static/%s", file))
	dst := fmt.Sprintf("%s\\tmp\\%s", pwd, file)
	if !util.PathExists(dst) {
		os.MkdirAll(fmt.Sprintf("%s\\tmp", pwd), 0740)
		destination, _ := os.Create(dst)
		defer destination.Close()
		io.Copy(destination, testFile)
		destination.Chmod(0777)
	}
	//解压文件
	dept, err := util.Unpack(context.Background(), dst, fmt.Sprintf("%s\\tmp\\", pwd))
	d, _ := os.Open(dept)
	d.Chmod(0777)
	log.Infof("dept=%s , err=%v", dept, err)
	log.Info("解压firefox完毕")
	return fmt.Sprintf("%s\\tmp\\firefox\\firefox.exe", pwd)
}
