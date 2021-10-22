# 使用说明

> ~~请安装 firefox浏览器~~
> 
> ~~使用的geckodrive是 v0.30.0版本~~
> 
> 请安装 Chrome浏览器
> 
> 使用的ChromeDriver 是 v95版本
> 
> 如果要使用其他版本，请到 `https://npm.taobao.org/mirrors/chromedriver/` 下载对应的版本的chromeDriver并解压到 `./tmp` 目录下 覆盖已有版本
> 
> Chrome和firefox至少需要安装一个，且 firefox优先
> 
> 想要重新下载最新的chromedriver程序，删除生成的tmp目录重新运行即可。
> 

## 第一次启动会加载一段时间，请耐心等待（下载electron依赖） 后续使用就快了。

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
$ go build -ldflags=" -s -X main.WebHookUrl=http://abc.com:8888/upsave" -o jd_ck_selenium

# 执行文件
$ ./jd_ck_selenium

# Windows requires special linker flags for GUI apps. 
# 添加了 `-H windowsgui` 后，windows上双击exe打开程序后不会直接弹出命令行窗口 
# 如果想要在windows下显示console页面，去掉 -H windowsgui 就行了。
$ go build -ldflags="-H windowsgui -s -X main.WebHookUrl=http://abc.com:8888/upsave" -o jd_ck_selenium.exe
```
## 详细的编译参数
> -ldflags="-X main.WebHookUrl=xxx" 这个不带也可以。就不进行cookie推送
> 
> 详细的 -ldflags样例 ：
> 
> ` -dflags=" -X main.WebHookUrl=http://abc.com:8888/upsave -X main.WebHookMethod=GET -X main.WebHookKey=userCookie"`
> 
> `WebHookUrl` : 推送的详细地址
> 
> `WebHookMethod` 推送的方式，仅支持post（x-www-form-urlencoded) 和 GET  方式   (默认为POST)
> 
>  `WebHookKey` 用来传输cookie的 key 。post的在body中，get的在url参数中。
>
> example样例: `go build -ldflags=" -s  -X main.WebHookUrl=http://abc.com:8888/upsave -X main.WebHookMethod=GET -X main.WebHookKey=userCookie"`
> 
## 编译demo
```shell
#!/bin/bash
GOOS=darwin GOARCH=amd64 go build -ldflags=" -s  -X main.WebHookUrl=http://aaa.bbb.ccc:7729/notify -X main.WebHookMethod=GET -X main.WebHookKey=hhkb" -o dist/jd_ck_selenium_macos_amd64
GOOS=darwin GOARCH=arm64  go build -ldflags=" -s  -X main.WebHookUrl=http://aaa.bbb.ccc:7729/notify -X main.WebHookMethod=GET -X main.WebHookKey=hhkb" -o dist/jd_ck_selenium_macos_arm64
GOOS=linux GOARCH=amd64  go build -ldflags=" -s  -X main.WebHookUrl=http://aaa.bbb.ccc:7729/notify -X main.WebHookMethod=GET -X main.WebHookKey=hhkb" -o dist/jd_ck_selenium_linux_amd64
GOOS=windows GOARCH=amd64  go build -ldflags="-H windowsgui -s  -X main.WebHookUrl=http://aaa.bbb.ccc:7729/notify -X main.WebHookMethod=GET -X main.WebHookKey=hhkb" -o dist/jd_ck_selenium_windows_amd64_noconsole.exe
GOOS=windows GOARCH=386  go build -ldflags="-H windowsgui -s  -X main.WebHookUrl=http://aaa.bbb.ccc:7729/notify -X main.WebHookMethod=GET -X main.WebHookKey=hhkb" -o dist/jd_ck_selenium_windows_i386_noconsole.exe
GOOS=windows GOARCH=amd64  go build -ldflags="-s  -X main.WebHookUrl=http://aaa.bbb.ccc:7729/notify -X main.WebHookMethod=GET -X main.WebHookKey=hhkb" -o dist/jd_ck_selenium_windows_amd64.exe
GOOS=windows GOARCH=386  go build -ldflags="-s  -X main.WebHookUrl=http://aaa.bbb.ccc:7729/notify -X main.WebHookMethod=GET -X main.WebHookKey=hhkb" -o dist/jd_ck_selenium_windows_i386.exe
```