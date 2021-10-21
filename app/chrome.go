package app

import (
	"embed"
	"errors"
	"fmt"
	"github.com/asticode/go-astilectron"
	"github.com/guonaihong/gout"
	"github.com/guonaihong/gout/dataflow"
	log "github.com/sirupsen/logrus"
	"github.com/tebeka/selenium"
	"go.uber.org/dig"
	"io"
	"jd_ck_selenium/util"
	"os"
	"runtime"
	"time"
)

//const cache_key_cookie = "CACHE_FOR_COOKIE_TOKEN_"

//var timeout = time.Second * 5

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
				ch.postWebHookCk(ct, cookie)
				return
			}
		}
		time.Sleep(time.Microsecond * 300)
	}
}

// 推送到远程服务器
func (ch *ChromeDriver) postWebHookCk(ct *dig.Container, cookie string) {
	// This will send a message and execute a callback
	// Callbacks are optional
	w.SendMessage(cookie, func(m *astilectron.EventMessage) {
		// Unmarshal
		var s string
		m.Unmarshal(&s)
		// Process message
		log.Printf("received %s\n", s)
	})
	var webhook WebHook
	////发送数据给 挂机服务器
	ct.Invoke(func(hook WebHook) {
		webhook = hook
	})
	postUrl := webhook.Url
	if postUrl != "" {
		var res MSG
		code := 0
		var flow *dataflow.DataFlow
		switch webhook.Method {
		case "GET":
			flow = gout.GET(webhook.Url).SetQuery(gout.H{
				webhook.Key: cookie,
			})
			break
		case "POST":
			flow = gout.POST(postUrl).SetWWWForm(
				gout.H{
					webhook.Key: cookie,
				},
			)
			break
		default:
			flow = gout.POST(postUrl)
			break
		}
		err := flow.
			BindJSON(&res).
			SetHeader(gout.H{
				"Connection":   "Keep-Alive",
				"Content-Type": "application/x-www-form-urlencoded; Charset=UTF-8",
				"Accept":       "application/json, text/plain, */*",
				"User-Agent":   "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/86.0.4240.111 Safari/537.36",
			}).
			Code(&code).
			SetTimeout(timeout).
			F().Retry().Attempt(5).
			WaitTime(time.Millisecond * 500).MaxWaitTime(time.Second * 5).
			Do()
		if err != nil || code != 200 {
			log.Errorf("upsave notify post  usercookie to %s faild", postUrl)
		} else {
			log.Infof("upsave to url %s post usercookie=%s success", postUrl, cookie)
		}
		return
	}
}

// 获取 系统和架构，读取geckodriver的位置
func (ch *ChromeDriver) GetChromeDriverPath(ct *dig.Container) (string, error) {
	path := "static/chromedriver-"
	osname := ""
	arch := ""
	var err error
	var f embed.FS
	ct.Invoke(func(static embed.FS) {
		f = static
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
		if osname == "win" {
			arch = "32.exe"
		}
		break
	case "386":
		arch = "32.exe"
		if osname != "win" {
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
	dst := "./chromedriver-" + osname + "-" + arch
	if osname == "win" {
		dst = dst + ".exe"
	}
	if !util.PathExists(dst) {
		destination, err := os.Create(dst)
		if err != nil {
			if osname != "win" {
				dst = "/tmp/chromedriver-" + osname + "-" + arch
				destination, err = os.Create(dst)
			} else {
				return "", err
			}
		}
		defer destination.Close()
		_, err = io.Copy(destination, testFile)
		destination.Chmod(0755)
	}
	return dst, err
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
