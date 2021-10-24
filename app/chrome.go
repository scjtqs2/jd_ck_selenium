package app

import (
	"context"
	"embed"
	"errors"
	"fmt"
	"github.com/guonaihong/gout"
	log "github.com/sirupsen/logrus"
	"github.com/tebeka/selenium"
	"github.com/tebeka/selenium/chrome"
	"go.uber.org/dig"
	"jd_ck_selenium/util"
	"os"
	"runtime"
	"strconv"
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
	util.DownloadSingle(context.Background(), src, fmt.Sprintf("%s/%s", dst, filename))
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
	//chromePath := ""
	chromePath := ch.checkChrome(ct)
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
	//chrome参数
	chromeCaps := chrome.Capabilities{
		Path: chromePath,
		MobileEmulation: &chrome.MobileEmulation{
			//DeviceName: "iPhone X",
			DeviceMetrics: &chrome.DeviceMetrics{
				Width:  375,
				Height: 812,
			},
			UserAgent: "Mozilla/5.0 (iPhone; CPU iPhone OS 14_6 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/14.0.3 Mobile/15E148 Safari/604.1",
		},
		Args: []string{
			//"--headless", // 设置Chrome无头模式，在linux下运行，需要设置这个参数，否则会报错
			//"--no-sandbox",
			"--user-agent=Mozilla/5.0 (iPhone; CPU iPhone OS 14_6 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/14.0.3 Mobile/15E148 Safari/604.1",
			"--window-size=375,812",
			"–-incognito",
			"--disable-infobars",
			"--start-maximized",
			"--no-sandbox",
			"--disable-gpu",
		},
	}
	caps.AddChrome(chromeCaps)
	ch.Wd, err = selenium.NewRemote(caps, fmt.Sprintf("http://localhost:%d/wd/hub", p))
	if err != nil {
		return err
	}
	//调整浏览器长宽高
	//ch.Wd.ResizeWindow("", 375, 812)

	// Navigate to the simple playground interface.
	if err = ch.Wd.Get("https://home.m.jd.com/myJd/newhome.action"); err != nil {
		return err
	}
	go ch.GetCookies(ct)
	return err
}

func (ch *ChromeDriver) CheckLastVersion() (version string, err error) {
	//if windowsOnly {
	//	return "2.45",nil
	//}
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
	version = "94.0.4606.61"
	log.Infof("latest chrome version =%s ", version)
	return version, err
}

//解压内置的firefox到tmp下去
func (ch *ChromeDriver) checkChrome(ct *dig.Container) string {

	var f embed.FS
	ct.Invoke(func(static embed.FS) {
		f = static
	})
	mirror := "https://ghproxy.com/https://github.com/scjtqs2/jd_ck_selenium/releases/download/v1.0.2/Chrome.zip"

	pwd := util.GetPwdPath()
	file := "Chrome.zip"
	dst := fmt.Sprintf("%s\\tmp\\%s", pwd, file)
	dprint := func(length, downLen int64) {
		//转float64
		size := float64(length)
		//下载大小转string
		Dstring := strconv.FormatInt(downLen, 10)
		//再转float64
		Dfloat, err := strconv.ParseFloat(Dstring, 64)

		if err != nil {
			log.Fatalln(err)
		}
		percent := util.Decimal(Dfloat / size * 100)
		//percentStr := strconv.FormatFloat(percent, 'f', -1, 64)
		percentStr := fmt.Sprintf("%.2f", percent)
		fmt.Printf("chrome下载中，已完成%s%s \n", percentStr, "%")
		//str="文件【"+filename+"】已下载了"+Dstring+"内容，总共有"+file_size+" 已完成:"+percentStr+"%"
	}
	if !util.PathExists(dst) {
		log.Info("下载chrome中，请稍等")
		//util.DownloadSingle(context.Background(), mirror, dst)
		util.DownloadMulit(mirror, dst, dprint)
		//wd := &sync.WaitGroup{}
		//wd.Add(1)
		//go util.DownloadFileBackend(mirror, dst, "", wd, dprint)
		//wd.Wait()
		log.Infof("下载chrome完成，准备解压")
	}

	//testFile, _ := f.Open(fmt.Sprintf("static/%s", file))

	//if !util.PathExists(dst) {
	//	// Makes sure destPath exists
	//	os.MkdirAll(fmt.Sprintf("%s\\tmp", pwd), 0740)
	//	destination, _ := os.Create(dst)
	//	defer destination.Close()
	//	io.Copy(destination, testFile)
	//	destination.Chmod(0777)
	//}
	log.Info("解压chrome中，请稍等")
	//解压文件
	dept, err := util.Unpack(context.Background(), dst, fmt.Sprintf("%s\\tmp\\", pwd))
	d, _ := os.Open(dept)
	d.Chmod(0777)
	if err != nil {
		log.Errorf("dept=%s , err=%v", dept, err)
	}
	log.Info("解压chrome完毕")
	return fmt.Sprintf("%s\\tmp\\Chrome\\App\\chrome.exe", pwd)
}
