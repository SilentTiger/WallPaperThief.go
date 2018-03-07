package logger

import (
	"log"
	"os"
)

const (
	prefixInfo  = "Info "
	prefixError = "Error "
)

var logger = log.New(os.Stdout, "", log.Ldate|log.Ltime)

// Info 打印普通日志
func Info(v ...interface{}) {
	logger.SetPrefix(prefixInfo)
	logger.Println(v)
}

// Error 打印错误日志
func Error(v ...interface{}) {
	logger.SetPrefix(prefixError)
	logger.Println(v)
}
