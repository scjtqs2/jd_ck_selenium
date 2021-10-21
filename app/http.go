package app

import (
	"embed"
	"fmt"
	"github.com/asticode/go-astilectron"
	"github.com/bluele/gcache"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/memstore"
	"github.com/gin-gonic/gin"
	"github.com/guonaihong/gout"
	"github.com/guonaihong/gout/dataflow"
	log "github.com/sirupsen/logrus"
	"go.uber.org/dig"
	"html/template"
	"net/http"
	"os"
	"time"
)

type httpServer struct {
	engine *gin.Engine
	HTTP   *http.Server
	ct     *dig.Container
}

type WebHook struct {
	Url    string
	Method string
	Key    string
}

var HTTPServer = &httpServer{}

var cache = gcache.New(20).LRU().Build()

const cache_key_cookie = "CACHE_FOR_COOKIE_TOKEN_"

var timeout = time.Second * 5

func (s *httpServer) httpStart(ct *dig.Container) int {
	var f embed.FS
	ct.Invoke(func(static embed.FS) {
		f = static
	})
	gin.SetMode(gin.ReleaseMode)
	s.engine = gin.New()
	// 创建基于 内存 的存储引擎，secret 参数是用于加密的密钥
	store := memstore.NewStore([]byte("scjtqsnb"))
	// 设置session中间件，参数mysession，指的是session的名字，也是cookie的名字
	// store是前面创建的存储引擎，我们可以替换成其他存储引擎
	s.engine.Use(sessions.Sessions("mysession", store))

	s.engine.Use(func(c *gin.Context) {
		if c.Request.Method != "GET" && c.Request.Method != "POST" {
			log.Warnf("已拒绝客户端 %v 的请求: 方法错误", c.Request.RemoteAddr)
			c.Status(404)
			return
		}
		c.Next()
	})
	// 自动加载模板
	t := template.New("tmp")
	//func 函数映射 全局模板可用
	t.Funcs(template.FuncMap{
		"getYear":        GetYear,
		"formatAsDate":   FormatAsDate,
		"getDate":        GetDate,
		"getavator":      Getavator,
		"getServerInfo":  GetServerInfo,
		"formatFileSize": FormatFileSize,
	})
	//从二进制中加载模板（后缀必须.html)
	templ := template.Must(template.New("").ParseFS(f, "static/www/html/*.html"))
	s.engine.SetHTMLTemplate(templ)
	//静态资源
	//s.engine.Static("/assets", "./template/assets")
	//s.engine.StaticFS("/public", http.FS(f))
	s.engine.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", gin.H{
			"version": 1.0,
		})
	})
	// 静态文件处理
	s.engine.GET("assets/*action", func(c *gin.Context) {
		c.FileFromFS("static/www/assets/"+c.Param("action"), http.FS(f))
	})
	port, _ := pickUnusedPort()
	//port := 10987
	go func() {
		s.HTTP = &http.Server{
			Addr:    fmt.Sprintf("127.0.0.1:%d", port),
			Handler: s.engine,
		}
		if err := s.HTTP.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Error(err)
			log.Infof("HTTP 服务启动失败, 请检查端口是否被占用.")
			log.Warnf("将在五秒后退出.")
			time.Sleep(time.Second * 5)
			os.Exit(1)
		}
	}()
	return port
}

// 推送到远程服务器
func postWebHookCk(ct *dig.Container, cookie string) {
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
