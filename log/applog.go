package log

import (
	"fmt"
	"io"
	"os"
	"pp/config"
	"sync"
	"time"
)

// 定义日志级别
const (
	LogDebug = 1 //调试日志
	LogInfo  = 2 //正常日志
	LogWarn  = 3 //警告
	LogError = 4 //错误
	LogFatal = 5 //严重错误
)

var (
	logger     *AppLogger
	loggerOnce sync.Once
)

type AppLogger struct {
	level          int
	logger         *Logger
	createTime     time.Time
	logFileMaxSize int64
}

func GetLogger() *AppLogger {
	loggerOnce.Do(func() {
		if logger == nil {
			logger = &AppLogger{}
		}
	})
	return logger
}

// 初始化日志系统
func (al *AppLogger) InitLogger() bool {
	appConf := config.NewAppConfig().GetConfig()
	timeStr := time.Now().Format("20060102150405")
	logFile, err := os.OpenFile(appConf.Logger+timeStr+".log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		fmt.Println("open file failed," + appConf.Logger)
		return false
	}
	al.logger = New(io.MultiWriter(logFile), "", Ldate|Lmicroseconds|Lshortfile)
	al.SetLevel(appConf.LoggerLevel)
	al.logFileMaxSize = appConf.LoggerFileMax
	if al.logFileMaxSize == 0 {
		al.logFileMaxSize = 1024 * 1024 * 100
	}
	al.createTime = time.Now()
	fileInfo, _ := logFile.Stat()
	al.logger.SetFileSize(fileInfo.Size())
	return true
}

// 更改日志文件，每个小时变更一次或者文件大小大于某个值
func (al *AppLogger) changeFile() {
	if time.Now().Hour() != al.createTime.Hour() || al.logger.GetFileSize() > al.logFileMaxSize {
		appConf := config.NewAppConfig().GetConfig()
		timeStr := time.Now().Format("20060102150405")
		logFile, err := os.OpenFile(appConf.Logger+timeStr+".log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			fmt.Println("open file failed," + appConf.Logger)
			return
		}
		al.logger.SetOutput(logFile)
		al.createTime = time.Now()
		al.logger.SetFileSize(0)
	}
}

func (al *AppLogger) SetLevel(level int) {
	al.level = level
}

func (al *AppLogger) SetLogFileMax(loggerFileMax int64) {
	al.logFileMaxSize = loggerFileMax
}

func (al *AppLogger) Debug(v ...interface{}) {
	if al.level > LogDebug {
		return
	}
	al.changeFile()
	al.logger.Println("[DEBUG]", fmt.Sprint(v...))
	//log.Println("[DEBUG]", fmt.Sprint(v...))
}

func (al *AppLogger) Info(v ...interface{}) {
	if al.level > LogInfo {
		return
	}
	al.changeFile()
	al.logger.Println("[INFO]", fmt.Sprint(v...))
	//log.Println("[INFO]", fmt.Sprint(v...))
}

func (al *AppLogger) Warn(v ...interface{}) {
	if al.level > LogWarn {
		return
	}
	al.changeFile()
	al.logger.Println("[WARN]", fmt.Sprint(v...))
	//log.Println("[WARN]", fmt.Sprint(v...))
}

func (al *AppLogger) Error(v ...interface{}) {
	if al.level > LogError {
		return
	}
	al.changeFile()
	al.logger.Println("[ERROR]", fmt.Sprint(v...))
	//log.Println("[ERROR]", fmt.Sprint(v...))
}

func (al *AppLogger) Fatal(v ...interface{}) {
	if al.level > LogFatal {
		return
	}
	al.changeFile()
	al.logger.Println("[FATAL]", fmt.Sprint(v...))
	//log.Println("[FATAL]", fmt.Sprint(v...))
}
