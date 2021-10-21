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
> 3、`go env -w CGO_ENABLED="0"` 关闭cgo
> 
> 4、 `go build` 即可得到jd_ck_selenium/jd_ck_selenium.exe 
>
> 5、添加推送地址
```bash
# Linux/macos
$ go build -ldflags=" -s -X main.WebHookUrl='http://abc.com:8888/upsave'" -o jd_ck_selenium

# 执行文件
$ ./jd_ck_selenium

# Windows requires special linker flags for GUI apps.
$ go build -ldflags="-H windowsgui -s -X main.WebHookUrl='http://abc.com:8888/upsave'" -o jd_ck_selenium.exe
```
## 详细的编译参数
> -ldflags="-X main.WebHookUrl=xxx" 这个不带也可以。就不进行cookie推送
> 
> 详细的 -ldflags样例 ：
> 
> ` -dflags=" -X main.WebHookUrl='http://abc.com:8888/upsave' -X main.WebHookMethod=GET -X main.WebHookKey=userCookie"`
> 
> `WebHookUrl` : 推送的详细地址
> 
> `WebHookMethod` 推送的方式，仅支持post（x-www-form-urlencoded) 和 GET  方式
> 
>  `WebHookKey` 用来传输cookie的 key 。post的在body中，get的在url参数中。
>
> example样例: `go build -ldflags=" -s  -X main.WebHookUrl='http://abc.com:8888/upsave' -X main.WebHookMethod=GET -X main.WebHookKey=userCookie"`