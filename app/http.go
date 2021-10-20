package app

import (
	"embed"
	"fmt"
	"github.com/bluele/gcache"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/memstore"
	"github.com/gin-gonic/gin"
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

var HTTPServer = &httpServer{}

var cache = gcache.New(20).LRU().Build()

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
	port ,_:=pickUnusedPort()
	go func() {
		s.HTTP = &http.Server{
			Addr:    fmt.Sprintf("127.0.0.1:%d",port),
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
