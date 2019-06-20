package logs

import (
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
	"time"
)

type LogType int

const (
	DEBUG    = LogType(iota) // 调试
	INFO                     // 重要
	NOTICE                   // 通知
	WARN                     // 警告
	ERROR                    // 错误
	CRITICAL                 // 严重
	FATAL                    // 致命
)

var LogTypeNames = func() []string {

	var (
		types = []string{
			"DEBUG",
			"INFO",
			"NOTICE",
			"WARN",
			"ERROR",
			"CRITICAL",
			"FATAL",
		}
		maxTypeNameLen = 0
	)

	for _, item := range types {
		if maxTypeNameLen < len(item) {
			maxTypeNameLen = len(item)
		}
	}

	for index, item := range types {
		typeLen := len(item)
		if typeLen < maxTypeNameLen {
			types[index] += strings.Repeat(" ", maxTypeNameLen-typeLen)
		}
	}

	return types
}()

// 								DEBUG| INFO| NOTICE | WARN |  ERROR | CRITICAL|FATAL
var logTypesColors = []string{"0;34", "0;32", "0;36", "1;33", "1;31", "1;31", "7;31"}

/**
日志
*/
type Logger struct {
	mn            sync.Mutex
	out           io.Writer
	logFormatFunc func(logType LogType, i interface{}) (string, []interface{}, bool)
	logLevel      LogType
}

/**
初始化日志
*/
func (logger *Logger) Init() {
	logger.mn.Lock()
	defer logger.mn.Unlock()
	logger.logFormatFunc = logger.DefaultLogFormatFunc
	logger.out = os.Stdout
	logger.logLevel = DEBUG
}

/**
设置日志级别
*/
func (logger *Logger) SetLogLevel(logType LogType) {
	logger.mn.Lock()
	defer logger.mn.Unlock()
	logger.logLevel = logType
}

/**
获取日志级别
*/
func (logger *Logger) GetLogLevel() LogType {
	logger.mn.Lock()
	defer logger.mn.Unlock()
	return logger.logLevel
}

/**
设置格式化log输出函数
*/
func (logger *Logger) setLoggerFormat(formatFunc func(logType LogType, i interface{}) (string, []interface{}, bool)) {
	logger.mn.Lock()
	defer logger.mn.Unlock()
	logger.logFormatFunc = formatFunc
}

func (logger *Logger) DefaultLogFormatFunc(logType LogType, i interface{}) (string, []interface{}, bool) {
	format := "\033[" + logTypesColors[logType] + "m%s [%s] %s \033[0m\n"
	layout := "2006-01-01 15:04:05.999"
	formatTime := time.Now().Format(layout)

	if len(formatTime) != len(layout) {
		//fmt.Println(len(layout) - len(formatTime))
		formatTime += ".000"[4-(len(layout)-len(formatTime)) : 4]
	}
	values := make([]interface{}, 3)
	values[0] = LogTypeNames[logType]
	values[1] = formatTime
	values[2] = fmt.Sprint(i)

	return format, values, true
}

func (logger *Logger) log(logType LogType, i interface{}) {
	logger.mn.Lock()
	defer logger.mn.Unlock()

	if logger.logLevel > logType {
		return
	}

	format, data, isLog := logger.logFormatFunc(logType, i)
	if !isLog {
		return
	}

	_, err := fmt.Fprintf(logger.out, format, data...)
	if err != nil {
		panic(err)
	}
}

// 输出信息
func (logger *Logger) Debug(i interface{}) {
	logger.log(DEBUG, i)
}
func (logger *Logger) Info(i interface{}) {
	logger.log(INFO, i)
}
func (logger *Logger) Notice(i interface{}) {
	logger.log(NOTICE, i)
}
func (logger *Logger) Warn(i interface{}) {
	logger.log(WARN, i)
}
func (logger *Logger) Error(i interface{}) {
	logger.log(ERROR, i)
}
func (logger *Logger) Critical(i interface{}) {
	logger.log(CRITICAL, i)
}
func (logger *Logger) Fatal(i interface{}) {
	logger.log(FATAL, i)
}

func NewLogger() *Logger {
	logger := new(Logger)
	logger.Init()
	return logger
}

func GetLogTypeNames(logType LogType) string {
	return LogTypeNames[logType]
}
