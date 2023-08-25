package utils

import (
	"fmt"
	"io"
	"os"

	"github.com/sirupsen/logrus"
	log "github.com/sirupsen/logrus"
	easy "github.com/t-tomalak/logrus-easy-formatter"
)

type Logger struct {
	log logrus.Logger
}

func (l *Logger) Setup() {
	l.log.SetLevel(log.WarnLevel)
	l.log.SetOutput(os.Stdout)
	l.log.Formatter = &logrus.TextFormatter{

		DisableColors:   true,
		TimestampFormat: "2006-01-02 15:04:05",
		FullTimestamp:   true,
	}
}

func newLogger() *Logger {
	logFile := "log.txt"
	f, err := os.OpenFile(logFile, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		fmt.Println("Failed to create logfile" + logFile)
		panic(err)
	}
	defer f.Close()
	log.SetOutput(f)
	log := &logrus.Logger{
		// Log into f file handler and on os.Stdout
		Out:   io.MultiWriter(f, os.Stdout),
		Level: logrus.DebugLevel,
		Formatter: &easy.Formatter{
			TimestampFormat: "2006-01-02 15:04:05",
			LogFormat:       "[%lvl%]: %time% - %msg%\n",
		},
	}

	l := &Logger{log: *log}
	return l
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
