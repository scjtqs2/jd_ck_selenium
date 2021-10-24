package util

import (
	"context"
	"errors"
	"fmt"
	"github.com/cavaliercoder/grab"
	log "github.com/sirupsen/logrus"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

var client = http.Client{Timeout: time.Second * 180}

var threadGroup = sync.WaitGroup{}

var packageSize int64

func init() {
	//每个线程下载文件的大小
	packageSize = 1048576 * 4
	//packageSize = 32 * 10240
}

// DownloadSingle is a cancellable function that downloads a src into a dst using a specific *http.Client and cleans up on
// failed downloads
func DownloadSingle(ctx context.Context, src, dst string) (err error) {
	// Log
	log.Debugf("Downloading %s into %s", src, dst)

	// Destination already exists
	if _, err = os.Stat(dst); err == nil {
		log.Debugf("%s already exists, skipping download...", dst)
		return
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("stating %s failed: %w", dst, err)
	}
	err = nil

	// Clean up on error
	defer func(err *error) {
		if *err != nil || ctx.Err() != nil {
			log.Debugf("Removing %s...", dst)
			os.Remove(dst)
		}
	}(&err)

	// Make sure the dst directory  exists
	if err = os.MkdirAll(filepath.Dir(dst), 0775); err != nil {
		return fmt.Errorf("mkdirall %s failed: %w", filepath.Dir(dst), err)
	}

	// create client
	client := grab.NewClient()
	req, _ := grab.NewRequest(dst, src)

	// start download
	fmt.Printf("Downloading %v...\n", req.URL())
	resp := client.Do(req)
	fmt.Printf("  %v\n", resp.HTTPResponse.Status)

	// start UI loop
	t := time.NewTicker(500 * time.Millisecond)
	defer t.Stop()

Loop:
	for {
		select {
		case <-t.C:
			fmt.Printf("  transferred %v / %v bytes (%.2f%%)\n",
				resp.BytesComplete(),
				resp.Size,
				100*resp.Progress())

		case <-resp.Done:
			// download is complete
			break Loop
		}
	}

	// check for errors
	if err = resp.Err(); err != nil {
		return fmt.Errorf("DownloadSingle failed: %w", err)
	}
	return nil
}

//DownloadSingle 多线程下载
func Download(url, cachePath string, scheduleCallback func(length, downLen int64)) (string, error) {
	// Make sure the dst directory  exists
	os.MkdirAll(filepath.Dir(cachePath), 0775)

	var localFileSize int64
	var file *os.File
	if info, e := os.Stat(cachePath); e != nil {
		if os.IsNotExist(e) {
			if createFile, err := os.Create(cachePath); err == nil {
				file = createFile
			} else {
				return "", err
			}
		} else {
			return "", e
		}
	} else {
		localFileSize = info.Size()
	}
	//HEAD 方法请求服务端是否支持多线程下载,并获取文件大小
	if request, e := http.NewRequest("HEAD", url, nil); e == nil {
		if response, i := client.Do(request); i == nil {
			defer response.Body.Close()
			//得到文件大小
			ContentLength := response.ContentLength
			if localFileSize == ContentLength {
				log.Warn("file exist~")
				return cachePath, nil
			} else {
				//判断是否支持多线下载
				defer func() {
					if err := recover(); err != nil {
						DownloadSingle(context.Background(), url, cachePath)
					}
				}()
				if strings.Compare(response.Header.Get("Accept-Ranges"), "bytes") == 0 {
					//支持 走下载流程
					if dispSliceDownload(file, ContentLength, url, scheduleCallback) == 0 {
						return cachePath, nil
					} else {
						err := DownloadSingle(context.Background(), url, cachePath)
						return cachePath, err
					}
				} else {
					//单线程下载
					err := DownloadSingle(context.Background(), url, cachePath)
					//panic("nonsupport ~")
					return cachePath, err
				}
			}
		} else {
			log.Errorf("HEAD faild err=%v", i)
			i = DownloadSingle(context.Background(), url, cachePath)
			return cachePath, i
			//panic(i)
		}
	} else {
		log.Errorf("HEAD faild err=%v", e)
		e = DownloadSingle(context.Background(), url, cachePath)
		return cachePath, e
	}
	return "", nil
}

func dispSliceDownload(file *os.File, ContentLength int64, url string, scheduleCallback func(length, downLen int64)) int {
	defer file.Close()
	//文件总大小除以 每个线程下载的大小
	i := ContentLength / packageSize
	//保证文件下载完整
	if ContentLength%packageSize > 0 {
		i += 1
	}
	//下载总进度
	var schedule int64
	//分配下载线程
	for count := 0; count < int(i); count++ {
		//计算每个线程下载的区间,起始位置
		var start int64
		var end int64
		start = int64(int64(count) * packageSize)
		end = start + packageSize
		if int64(end) > ContentLength {
			end = end - (end - ContentLength)
		}
		//构建请求
		if req, e := http.NewRequest("GET", url, nil); e == nil {
			req.Header.Set(
				"Range",
				"bytes="+strconv.FormatInt(start, 10)+"-"+strconv.FormatInt(end, 10))
			//
			threadGroup.Add(1)
			go sliceDownload(req, file, &schedule, &ContentLength, scheduleCallback, start)
		} else {
			panic(e)
		}

	}
	//等待所有线程完成下载
	threadGroup.Wait()
	return 0
}

func sliceDownload(request *http.Request, file *os.File, schedule *int64, ContentLength *int64, scheduleCallback func(length, downLen int64),
	start int64) {
	defer threadGroup.Done()
	if response, e := client.Do(request); e == nil && response.StatusCode == 206 {
		defer response.Body.Close()
		if bytes, i := ioutil.ReadAll(response.Body); i == nil {
			i2 := len(bytes)
			//从我们计算好的起点写入文件
			file.WriteAt(bytes, start)
			atomic.AddInt64(schedule, int64(i2))
			val := atomic.LoadInt64(schedule)
			//num := float64(val*1.0) / float64(*ContentLength) * 100
			scheduleCallback(val,*ContentLength)
		} else {
			panic(e)
		}
	} else {
		log.Errorf("error =%v,response =%+v", e, response)
		panic(fmt.Sprintf("error =%v,response =%+v", e, response))
	}
}

//通过gorouting 后台下载
func DownloadFileBackend(url string, localPath string, cookies string,wd *sync.WaitGroup, fb func(length, downLen int64)) error {
	var (
		fsize   int64
		buf     = make([]byte, 32*1024)
		written int64
	)
	defer wd.Done()
	// Make sure the dst directory  exists
	os.MkdirAll(filepath.Dir(localPath), 0775)

	tmpFilePath := localPath + ".download"
	//fmt.Println(tmpFilePath)
	//创建一个http client
	//client := new(http.Client)
	//client.Timeout = time.Second * 60 //设置超时时间
	client := &http.Client{}
	//提交请求
	reqest, err := http.NewRequest("GET", url, nil)
	//增加header选项
	if cookies != "" {
		reqest.Header.Add("Cookie", cookies)
	}
	reqest.Header.Add("User-Agent", "Mozilla/5.0 (Windows NT 6.1; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/75.0.3770.100 Safari/537.36")
	//reqest.Header.Add("X-Requested-With", "xxxx")
	//get方法获取资源
	resp, err := client.Do(reqest)
	if err != nil {
		return err
	}

	//读取服务器返回的文件大小
	fsize, err = strconv.ParseInt(resp.Header.Get("Content-Length"), 10, 32)
	if err != nil {
		log.Error(err)
	}
	if IsFileExist(localPath, fsize) {
		return err
	}
	log.Infof("文件总大小：%d", fsize)
	//创建文件
	file, err := os.Create(tmpFilePath)
	if err != nil {
		return err
	}
	defer file.Close()
	if resp.Body == nil {
		return errors.New("body is null")
	}
	defer resp.Body.Close()
	//下面是 io.copyBuffer() 的简化版本
	for {
		//读取bytes
		nr, er := resp.Body.Read(buf)
		if nr > 0 {
			//写入bytes
			nw, ew := file.Write(buf[0:nr])
			//数据长度大于0
			if nw > 0 {
				written += int64(nw)
			}
			//写入出错
			if ew != nil {
				err = ew
				break
			}
			//读取是数据长度不等于写入的数据长度
			if nr != nw {
				err = io.ErrShortWrite
				break
			}
		}
		if er != nil {
			if er != io.EOF {
				err = er
			}
			break
		}
		//没有错误了快使用 callback
		fb(fsize, written)
	}
	log.Errorf("下载出错 err=%v", err)
	if err == nil {
		file.Close()
		err = os.Rename(tmpFilePath, localPath)
		log.Errorf("rename出错 err=%v", err)
	}
	return err
}

func IsFileExist(filename string, filesize int64) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		fmt.Println(info)
		return false
	}
	if filesize == info.Size() {
		fmt.Println("file exist ！", info.Name(), info.Size(), info.ModTime())
		return true
	}
	del := os.Remove(filename)
	if del != nil {
		fmt.Println(del)
	}
	return false
}
