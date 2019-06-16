package logs

import (
	"fmt"
	"os"
	"runtime/debug"
	"time"
)

type RotateFileLogger struct {
	Logger
	file               *os.File                    // 正在操作的文件
	dirPath            string                      // logs文件所在的目录
	fileNameFormatFunc func(time time.Time) string // 获取文件名格式
	newFileGapTime     time.Duration               // 创建新log的时间间隔
	lastFileTime       time.Time                   // 上一次创建文件的时间
}

func (fileLogger *RotateFileLogger) Init(dir string) {
	fileLogger.Logger.Init()
	fileLogger.mn.Lock()
	defer fileLogger.mn.Unlock()

	fileLogger.fileNameFormatFunc = fileLogger.DefaultFileNameFormat
	fileLogger.logFormatFunc = fileLogger.DefaultLogFormatFunc
	fileLogger.newFileGapTime = 0
	fileLogger.lastFileTime = time.Now()
	fileLogger.dirPath = dir

	f, err := fileLogger.createLogFile(fileLogger.fileNameFormatFunc(fileLogger.lastFileTime))
	if err != nil {
		panic(err)
		return
	}

	fileLogger.file = f
	fileLogger.out = f
}

func (fileLogger *RotateFileLogger) DefaultFileNameFormat(fileName time.Time) string {
	layout := "2006-01-01 15-04-05.999"
	formatTime := fileName.Format(layout)
	if len(formatTime) != len(layout) {
		formatTime += ".000"[4-(len(layout)-len(formatTime)) : 4]
	}
	return fmt.Sprintf("%s.log", formatTime)
}

func (fileLogger *RotateFileLogger) createLogFile(fileName string) (f *os.File, err error) {
	if len(fileLogger.dirPath) != 0 {
		fileName = fmt.Sprintf("%s/%s", fileLogger.dirPath, fileName)
		err = os.MkdirAll(fileLogger.dirPath, 0777)
		if err != nil {
			panic(err)
		}
	}
	f, err = os.OpenFile(fileName, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0666)
	if err != nil {
		f = nil
		return
	}
	err = nil
	return
}

func (fileLogger *RotateFileLogger) DefaultLogFormatFunc(logType LogType, i interface{}) (string, []interface{}, bool) {
	format := "%s [%s] %s\n"
	now := time.Now()
	gapTime := now.Sub(fileLogger.lastFileTime)
	if gapTime > fileLogger.newFileGapTime && fileLogger.newFileGapTime > 0 {
		fileLogger.file.Close()
		rate := int(int64(gapTime) / int64(fileLogger.newFileGapTime))
		fileLogger.lastFileTime = fileLogger.lastFileTime.Add(fileLogger.newFileGapTime * time.Duration(rate))
		file, err := fileLogger.createLogFile(fileLogger.fileNameFormatFunc(fileLogger.lastFileTime))
		if err != nil {
			file.Close()
			panic(err)
		}
		fileLogger.file = file
		fileLogger.out = file
	}

	layout := "2006-01-01 15:04:05.999"
	formatTime := now.Format(layout)
	defer func() {
		e := recover()
		if e != nil {
			panic(debug.Stack())
		}
	}()
	if len(formatTime) != len(layout) {
		formatTime += ".000"[4-(len(layout)-len(formatTime)) : 4]
	}

	values := make([]interface{}, 3)
	values[0] = GetLogTypeNames(logType)
	values[1] = formatTime
	values[2] = fmt.Sprint(i)

	return format, values, true
}

func (fileLogger *RotateFileLogger) SetNewFileGapTime(gapTime time.Duration) {
	fileLogger.mn.Lock()
	defer fileLogger.mn.Unlock()
	fileLogger.newFileGapTime = gapTime
}

func NewRotateFileLogger(dir string) *RotateFileLogger {
	fileLogger := new(RotateFileLogger)
	fileLogger.Init(dir)
	return fileLogger
}
