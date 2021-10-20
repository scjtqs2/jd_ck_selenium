package util

import (
	"errors"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

// WriteAllText 将给定text写入给定path
func WriteAllText(path, text string) error {
	return ioutil.WriteFile(path, []byte(text), 0o644)
}

// GetPwdPath 获取当前文件的路径，直接返回string
func GetPwdPath() string {
	cwd, e := os.Getwd()
	if e != nil {
		panic(e)
	}
	return cwd
}

//获取当前真路径
func GetCurrentPath() (string, error) {
	file, err := exec.LookPath(os.Args[0])
	if err != nil {
		return "", err
	}
	path, err := filepath.Abs(file)
	if err != nil {
		return "", err
	}
	if runtime.GOOS == "windows" {
		path = strings.Replace(path, "\\", "/", -1)
	}
	i := strings.LastIndex(path, "/")
	if i < 0 {
		return "", errors.New(`system/path_error Can't find "/" or "\".`)
	}
	return string(path[0 : i+1]), nil
}

// PathExists 判断给定path是否存在
func PathExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil || os.IsExist(err)
}

// ReadAllText 读取给定path对应文件，无法读取时返回空值
func ReadAllText(path string) string {
	b, err := ioutil.ReadFile(path)
	if err != nil {
		log.Error(err)
		return ""
	}
	return string(b)
}