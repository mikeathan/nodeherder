package utils

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"

	"github.com/sirupsen/logrus"
	log "github.com/sirupsen/logrus"
	easy "github.com/t-tomalak/logrus-easy-formatter"
)

var logger *Logger
var once sync.Once

const LogPath string = "logs"

func GetInstance() *Logger {
	once.Do(func() {
		logger = newFileLogger("nodeherder.log")
	})
	return logger
}

type Logger struct {
	log  *logrus.Logger
	file *os.File
}

func newConsoleLogger() *Logger {

	log := &logrus.Logger{
		Out:   os.Stdout,
		Level: logrus.InfoLevel,
		Formatter: &easy.Formatter{
			TimestampFormat: "2006-01-02 15:04:05",
			LogFormat:       "[%lvl%]: %time% - %msg%\n",
		},
	}

	return &Logger{log: log}
}

func createDirIfNotExists() {
	if _, err := os.Stat(LogPath); errors.Is(err, os.ErrNotExist) {
		err := os.Mkdir(LogPath, os.ModePerm)
		if err != nil {
			log.Println(fmt.Sprintf("Failed to create log directory %s Error: %v", LogPath, err))
		}
	}
}

func newFileLogger(logName string) *Logger {

	createDirIfNotExists()
	f, err := os.OpenFile(filepath.Join(LogPath, logName), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		fmt.Println("Failed to create logfile" + err.Error())
		panic(err)
	}

	log := &logrus.Logger{
		Out:   io.MultiWriter(f, os.Stdout),
		Level: logrus.InfoLevel,
		Formatter: &easy.Formatter{
			TimestampFormat: "2006-01-02 15:04:05",
			LogFormat:       "[%lvl%]: %time% - %msg%\n",
		},
	}

	return &Logger{log: log, file: f}
}

func (l *Logger) SetLevel(level string) {
	ll, err := log.ParseLevel(level)
	if err != nil {
		l.Errorf("undefined level %s\n", level)
		return
	}

	l.log.SetLevel(ll)
	l.Infof("set level: %s\n", ll.String())
}

func (l *Logger) Close() {
	if l.file != nil {
		logger.Info("file logger disposed")
		l.file.Close()
	}
}

func (l *Logger) Debug(msg ...interface{}) {
	l.log.Debug(msg...)
}

func (l *Logger) Info(msg ...interface{}) {
	l.log.Info(msg...)
}

func (l *Logger) Warning(msg ...interface{}) {
	l.log.Warning(msg...)
}

func (l *Logger) Error(msg ...interface{}) {
	l.log.Error(msg...)
}

func (l *Logger) Debugf(format string, msg ...interface{}) {
	l.log.Debugf(format, msg...)
}

func (l *Logger) Infof(format string, msg ...interface{}) {
	l.log.Infof(format, msg...)
}

func (l *Logger) Warningf(format string, msg ...interface{}) {
	l.log.Warnf(format, msg...)
}

func (l *Logger) Errorf(format string, msg ...interface{}) {
	l.log.Errorf(format, msg...)
}
