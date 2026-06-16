package service

import (
	"log"
	"os"
)

// Logger 全局日志实例
var Logger = log.New(os.Stdout, "[API] ", log.LstdFlags|log.Lshortfile)

// LogError 记录错误日志
func LogError(msg string, err error) {
	if err != nil {
		Logger.Printf("ERROR: %s: %v", msg, err)
	}
}

// LogInfo 记录信息日志
func LogInfo(msg string) {
	Logger.Printf("INFO: %s", msg)
}

// LogWarn 记录警告日志
func LogWarn(msg string) {
	Logger.Printf("WARN: %s", msg)
}
