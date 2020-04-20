package utils

import (
	"os"
	"path/filepath"
)

// API目录常量
const ApiDir = "api"

// GetCurrentPath获取运行程序绝对路径，如：/Users/wumoxi/dev/go/src/hello-demo
func GetCurrentPath() string {
	cur, _ := os.Getwd()
	return cur
}

// GetCurrentDir获取路径最后一级目录名称, 如：/Users/wumoxi/dev/go/src/hello-demo -> hello-demo
func GetCurrentDir(path string) string {
	_, file := filepath.Split(path)
	return file
}
