package utils

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/sirupsen/logrus"
	easy "github.com/t-tomalak/logrus-easy-formatter"
)

const LogPath string = "logs"

var log *logger = newConsoleLogger()

func InitFileLogger() {
	if log != nil {
		log.Close()
	}

	log = newFileLogger("nodeherder.log")
}

type logger struct {
	log  *logrus.Logger
	file *os.File
}

func newConsoleLogger() *logger {

	log := &logrus.Logger{
		Out:   os.Stdout,
		Level: logrus.InfoLevel,
		Formatter: &easy.Formatter{
			TimestampFormat: "2006-01-02 15:04:05",
			LogFormat:       "[%lvl%]: %time% - %msg%\n",
		},
	}

	return &logger{log: log}
}

func createDirIfNotExists() {
	if _, err := os.Stat(LogPath); errors.Is(err, os.ErrNotExist) {
		err := os.Mkdir(LogPath, os.ModePerm)
		if err != nil {
			fmt.Println(fmt.Sprintf("Failed to create log directory %s Error: %v", LogPath, err))
			panic(err)
		}
	}
}

func newFileLogger(logName string) *logger {

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

	return &logger{log: log, file: f}
}

func SetLogLevel(level string) {
	log.SetLevel(level)
}

func Close() {
	log.Close()
}

func LogDebug(msg ...interface{}) {
	log.Debug(msg...)
}

func LogInfo(msg ...interface{}) {
	log.Info(msg...)
}

func LogWarn(msg ...interface{}) {
	log.Warn(msg...)
}

func LogError(msg ...interface{}) {
	log.Error(msg...)
}

func LogDebugf(format string, msg ...interface{}) {
	log.Debugf(format, msg...)
}

func LogInfof(format string, msg ...interface{}) {
	log.Infof(format, msg...)
}

func LogWarnf(format string, msg ...interface{}) {
	log.Warnf(format, msg...)
}

func LogErrorf(format string, msg ...interface{}) {
	log.Errorf(format, msg...)
}

func (l *logger) SetLevel(level string) {
	ll, err := logrus.ParseLevel(level)
	if err != nil {
		l.Errorf("undefined level %s\n", level)
		return
	}

	l.log.SetLevel(ll)
	l.Infof("set level: %s\n", ll.String())
}

func (l *logger) Close() {
	if l.file != nil {
		l.Info("file logger disposed")
		l.file.Close()
	}
}

func (l *logger) Debug(msg ...interface{}) {
	l.log.Debug(msg...)
}

func (l *logger) Info(msg ...interface{}) {
	l.log.Info(msg...)
}

func (l *logger) Warn(msg ...interface{}) {
	l.log.Warning(msg...)
}

func (l *logger) Error(msg ...interface{}) {
	l.log.Error(msg...)
}

func (l *logger) Debugf(format string, msg ...interface{}) {
	l.log.Debugf(format, msg...)
}

func (l *logger) Infof(format string, msg ...interface{}) {
	l.log.Infof(format, msg...)
}

func (l *logger) Warnf(format string, msg ...interface{}) {
	l.log.Warnf(format, msg...)
}

func (l *logger) Errorf(format string, msg ...interface{}) {
	l.log.Errorf(format, msg...)
}
