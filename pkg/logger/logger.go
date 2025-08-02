package logger

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"

	"github.com/sirupsen/logrus"
)

var (
	logger *logrus.Logger
)

// Init 初始化日志
func Init(level string, logFile string) error {
	logger = logrus.New()

	// 设置日志级别
	switch level {
	case "debug":
		logger.SetLevel(logrus.DebugLevel)
	case "info":
		logger.SetLevel(logrus.InfoLevel)
	case "warn":
		logger.SetLevel(logrus.WarnLevel)
	case "error":
		logger.SetLevel(logrus.ErrorLevel)
	case "fatal":
		logger.SetLevel(logrus.FatalLevel)
	default:
		logger.SetLevel(logrus.InfoLevel)
	}

	// 设置日志格式
	logger.SetFormatter(&logrus.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: "2006-01-02 15:04:05",
		ForceColors:     true,
	})

	// 设置输出
	if logFile != "" {
		// 确保日志目录存在
		logDir := filepath.Dir(logFile)
		if err := os.MkdirAll(logDir, 0755); err != nil {
			return fmt.Errorf("failed to create log directory: %v", err)
		}

		// 打开日志文件
		file, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			return fmt.Errorf("failed to open log file: %v", err)
		}

		// 同时输出到文件和控制台
		mw := io.MultiWriter(os.Stdout, file)
		logger.SetOutput(mw)
	} else {
		logger.SetOutput(os.Stdout)
	}

	return nil
}

// Debug 输出调试日志（受调试总开关控制）
func Debug(args ...interface{}) {
	if shouldLogDebug() {
		if logger != nil {
			logger.Debug(args...)
		} else {
			log.Println(args...)
		}
	}
}

// Debugf 输出格式化调试日志（受调试总开关控制）
func Debugf(format string, args ...interface{}) {
	if shouldLogDebug() {
		if logger != nil {
			logger.Debugf(format, args...)
		} else {
			log.Printf(format, args...)
		}
	}
}

// Info 输出信息日志
func Info(args ...interface{}) {
	if logger != nil {
		logger.Info(args...)
	} else {
		log.Println(args...)
	}
}

// Infof 输出格式化信息日志
func Infof(format string, args ...interface{}) {
	if logger != nil {
		logger.Infof(format, args...)
	} else {
		log.Printf(format, args...)
	}
}

// Warn 输出警告日志
func Warn(args ...interface{}) {
	if logger != nil {
		logger.Warn(args...)
	} else {
		log.Println(args...)
	}
}

// Warnf 输出格式化警告日志
func Warnf(format string, args ...interface{}) {
	if logger != nil {
		logger.Warnf(format, args...)
	} else {
		log.Printf(format, args...)
	}
}

// Error 输出错误日志
func Error(args ...interface{}) {
	if logger != nil {
		logger.Error(args...)
	} else {
		log.Println(args...)
	}
}

// Errorf 输出格式化错误日志
func Errorf(format string, args ...interface{}) {
	if logger != nil {
		logger.Errorf(format, args...)
	} else {
		log.Printf(format, args...)
	}
}

// Fatal 输出致命错误日志并退出
func Fatal(args ...interface{}) {
	if logger != nil {
		logger.Fatal(args...)
	} else {
		log.Fatal(args...)
	}
}

// Fatalf 输出格式化致命错误日志并退出
func Fatalf(format string, args ...interface{}) {
	if logger != nil {
		logger.Fatalf(format, args...)
	} else {
		log.Fatalf(format, args...)
	}
}

// Log 输出指定级别的日志
func Log(level string, message string) {
	switch level {
	case "DEBUG":
		Debug(message)
	case "INFO":
		Info(message)
	case "WARN":
		Warn(message)
	case "ERROR":
		Error(message)
	case "FATAL":
		Fatal(message)
	default:
		Info(message)
	}
}

// SetLevel 设置日志级别
func SetLevel(level string) {
	if logger == nil {
		return
	}

	switch level {
	case "debug":
		logger.SetLevel(logrus.DebugLevel)
	case "info":
		logger.SetLevel(logrus.InfoLevel)
	case "warn":
		logger.SetLevel(logrus.WarnLevel)
	case "error":
		logger.SetLevel(logrus.ErrorLevel)
	case "fatal":
		logger.SetLevel(logrus.FatalLevel)
	}
}

// GetLogger 获取日志实例
func GetLogger() *logrus.Logger {
	return logger
}

// shouldLogDebug 检查是否应该输出调试日志
func shouldLogDebug() bool {
	// 检查调试总开关
	debugEnabled := os.Getenv("DEBUG_ENABLED")
	if debugEnabled == "true" {
		return true
	}

	// 检查详细日志开关
	verboseLogging := os.Getenv("VERBOSE_LOGGING")
	if verboseLogging == "true" {
		return true
	}

	// 检查日志级别
	logLevel := os.Getenv("LOG_LEVEL")
	if logLevel == "debug" {
		return true
	}

	return false
}
