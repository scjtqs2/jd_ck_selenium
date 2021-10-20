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
# Linux/macos
$ go build -ldflags " -s -X main.WebHookUrl='http://abc.com:8888/upsave'" -o jd_ck_selenium && ./jd_ck_selenium

# Windows requires special linker flags for GUI apps.

$ go build -ldflags="-H windowsgui -s -X main.WebHookUrl='http://abc.com:8888/upsave'" -o jd_ck_selenium.exe
```
> -X main.WebHookUrl 这个不带也可以。就不进行cookie推送
