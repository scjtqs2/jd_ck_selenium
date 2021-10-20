# 使用说明

> 请安装 firefox浏览器
> 
> 使用的geckodrive是 v0.30.0版本
> 

## 编译方法
> 1、安装 golang 1.16+版本
> 
> 2、`go env -w GOPROXY="http://goproxy.cn,direct"` 设置代理
>  
> 3、`go env -w CGO_ENABLED="1"` 开启cgo
> 
> 4、 `go build` 即可得到jd_ck_selenium/jd_ck_selenium.exe 
>

```bash
# Linux
$ go build -ldflags " -s -X main.WebHookUrl='http://abc.com:8888/upsave'" -o jd_ck_selenium && ./jd_ck_selenium

# MacOS uses app bundles for GUI apps
$ mkdir -p example.app/Contents/MacOS
$ go build " -s -X main.WebHookUrl='http://abc.com:8888/upsave'" -o example.app/Contents/MacOS/example
$ open example.app # Or click on the app in Finder

# Windows requires special linker flags for GUI apps.
# It's also recommended to use TDM-GCC-64 compiler for CGo.
# http://tdm-gcc.tdragon.net/download
$ go build -ldflags="-H windowsgui -s -X main.WebHookUrl='http://abc.com:8888/upsave'" -o jd_ck_selenium.exe
```

## windows下 提示缺失 webview.dll
> 下载dll目录下的对应架构的 两个dll到 exe的当前目录
