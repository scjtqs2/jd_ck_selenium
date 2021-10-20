package app

import (
	"embed"
	"errors"
	"fmt"
	"github.com/guonaihong/gout"
	log "github.com/sirupsen/logrus"
	"github.com/tebeka/selenium"
	"go.uber.org/dig"
	"io"
	"jd_ck_selenium/util"
	"os"
	"runtime"
	"time"
)

const cache_key_cookie = "CACHE_FOR_COOKIE_TOKEN_"

var timeout = time.Second * 5

// 获取cookie并校验cookie 是否存在
func GetCookies(wd selenium.WebDriver) {
	for {
		select {
		case <-c:
			return
		default:
			cks, err := wd.GetCookies()
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
				return
			}
		}
		time.Sleep(time.Microsecond * 300)
	}
}

// 推送到远程服务器
func postWebHookCk(ct *dig.Container, cookie string) {
	////发送数据给 挂机服务器
	postUrl := ""
	ct.Invoke(func(WebHookUrl string) {
		postUrl = WebHookUrl
	})
	if postUrl != "" {
		var res MSG
		code := 0
		err := gout.POST(postUrl).
			//Debug(true).
			SetWWWForm(
				gout.H{
					"userCookie": cookie,
				},
			).
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
func GetGeckoDriverPath(ct *dig.Container) (string, error) {
	path := "static/geckodriver-"
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
			arch = "amd64.exe"
		}
		break
	case "386":
		arch = "i386.exe"
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
	dst := "./geckodriver-" + osname + "-" + arch
	if osname == "win" {
		dst = dst + ".exe"
	}
	if !util.PathExists(dst) {
		destination, err := os.Create(dst)
		if err != nil {
			if osname != "win" {
				dst = "/tmp/geckodriver-" + osname + "-" + arch
				destination, err = os.Create(dst)
			} else {
				return "", err
			}
		}
		defer destination.Close()
		_, err = io.Copy(destination, testFile)
		destination.Chmod(0755)
	}
	copydll(ct)
	return dst, err
}

func copydll(ct *dig.Container)  {
	if runtime.GOOS == "windows" {
		var f embed.FS
		ct.Invoke(func(static embed.FS) {
			f = static
		})
		d1 :="static/webview.dll"
		d2 :="static/WebView2Loader.dll"
		sd1 := "./webview.dll"
		sd2 := "./WebView2Loader.dl"
		if !util.PathExists(sd1) {
			testFile, err := f.Open(d1)
			destination, err := os.Create(sd1)
			if err != nil {
				return
			}
			defer destination.Close()
			_, err = io.Copy(destination, testFile)
		}
		if !util.PathExists(sd2) {
			testFile, err := f.Open(d2)
			destination, err := os.Create(sd1)
			if err != nil {
				return
			}
			defer destination.Close()
			_, err = io.Copy(destination, testFile)
		}
	}
}

func seRun(ct *dig.Container) {
	p, _ := pickUnusedPort()
	opts := []selenium.ServiceOption{
		//selenium.StartFrameBuffer(),           // Start an X frame buffer for the browser to run in.
		//selenium.Output(os.Stderr), // Output debug information to STDERR.
	}
	var err error
	//defer func() {
	//	c <- os.Kill
	//}()
	geckoDriverPath, err = GetGeckoDriverPath(ct)
	if err != nil {
		panic(err)
	}
	selenium.SetDebug(false)
	service, err = selenium.NewGeckoDriverService(geckoDriverPath, p, opts...)
	if err != nil {
		panic(err) // panic is used only as an example and is not otherwise recommended.
	}

	// Connect to the WebDriver instance running locally.
	caps := selenium.Capabilities{"browserName": "firefox"}
	wd, err = selenium.NewRemote(caps, fmt.Sprintf("http://localhost:%d", p))
	if err != nil {
		panic(err)
	}

	// Navigate to the simple playground interface.
	if err := wd.Get("https://home.m.jd.com/myJd/newhome.action"); err != nil {
		panic(err)
	}
	go GetCookies(wd)
}
