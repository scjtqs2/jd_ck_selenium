package app

import (
	"fmt"
	"github.com/google/uuid"
	"net"
	"os"
	"runtime"
	"time"
)

type MSG map[string]interface{}

//格式化年月日
func FormatAsDate(t time.Time) string {
	year, month, day := t.Date()
	return fmt.Sprintf("%d%02d/%02d", year, month, day)
}

// 获取年份
func GetYear() string {
	t := time.Now()
	year, _, _ := t.Date()
	return fmt.Sprintf("%d", year)
}

// 获取当前年月日
func GetDate() string {
	t := time.Now()
	year, month, day := t.Date()
	return fmt.Sprintf("%d-%02d-%02d", year, month, day)
}

// 随机获取一个头像
func Getavator() string {
	Uuid := uuid.New().String()
	grav_url := "https://www.gravatar.com/avatar/" + Uuid
	return grav_url
}

type info struct {
	Root       string
	Version    string
	Hostname   string
	Interfaces interface{}
	Goarch     string
	Goos       string
	//VirtualMemory *mem.VirtualMemoryStat
	Sys         uint64
	CpuInfoStat struct {
		Count   int
		Percent []float64
	}
}

func GetServerInfo() *info {
	root := runtime.GOROOT()          // GO 路径
	version := runtime.Version()      //GO 版本信息
	hostname, _ := os.Hostname()      //获得PC名
	interfaces, _ := net.Interfaces() //获得网卡信息
	goarch := runtime.GOARCH          //系统构架 386、amd64
	goos := runtime.GOOS              //系统版本 windows
	Info := &info{
		Root:       root,
		Version:    version,
		Hostname:   hostname,
		Interfaces: interfaces,
		Goarch:     goarch,
		Goos:       goos,
	}

	//v, _ := mem.VirtualMemory()
	//Info.VirtualMemory = v
	var ms runtime.MemStats
	runtime.ReadMemStats(&ms)
	Info.Sys = ms.Sys
	//Info.CpuInfoStat.Count, _ = cpu.Counts(true)
	//Info.CpuInfoStat.Percent, _ = cpu.Percent(0, true)
	return Info
}

// 字节的单位转换 保留两位小数
func FormatFileSize(fileSize uint64) (size string) {
	if fileSize < 1024 {
		//return strconv.FormatInt(fileSize, 10) + "B"
		return fmt.Sprintf("%.2fB", float64(fileSize)/float64(1))
	} else if fileSize < (1024 * 1024) {
		return fmt.Sprintf("%.2fKB", float64(fileSize)/float64(1024))
	} else if fileSize < (1024 * 1024 * 1024) {
		return fmt.Sprintf("%.2fMB", float64(fileSize)/float64(1024*1024))
	} else if fileSize < (1024 * 1024 * 1024 * 1024) {
		return fmt.Sprintf("%.2fGB", float64(fileSize)/float64(1024*1024*1024))
	} else if fileSize < (1024 * 1024 * 1024 * 1024 * 1024) {
		return fmt.Sprintf("%.2fTB", float64(fileSize)/float64(1024*1024*1024*1024))
	} else { //if fileSize < (1024 * 1024 * 1024 * 1024 * 1024 * 1024)
		return fmt.Sprintf("%.2fEB", float64(fileSize)/float64(1024*1024*1024*1024*1024))
	}
}
