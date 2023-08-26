package utils

import (
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"

	"github.com/sirupsen/logrus"
	easy "github.com/t-tomalak/logrus-easy-formatter"
)

const LogPath string = "logs"

var filelogger = newFileLogger("nodeherder.log")

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
			log.Println(fmt.Sprintf("Failed to create log directory %s Error: %v", LogPath, err))
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

func (l *logger) Warning(msg ...interface{}) {
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

func (l *logger) Warningf(format string, msg ...interface{}) {
	l.log.Warnf(format, msg...)
}

func (l *logger) Errorf(format string, msg ...interface{}) {
	l.log.Errorf(format, msg...)
}
