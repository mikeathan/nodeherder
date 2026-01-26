package utils

import (
	"encoding/json"
	"io"
	"node-herder/models/logging"
	"os"
	"path/filepath"

	"github.com/sirupsen/logrus"
	easy "github.com/t-tomalak/logrus-easy-formatter"
	"gopkg.in/natefinch/lumberjack.v2"
)

const LogsPath string = "logs"
const LogName string = "nodeherder.log"
const LogExtension = ".log"

var log *logger = newConsoleLogger()
var remoteHook *RemoteHook = newRemoteHook()

// InitFileLogger initializes the logger with file output.
// Logs are written to both a file and stderr (Unix convention).
// stdout is reserved for program output only.
func InitFileLogger() {
	log = newFileLogger(LogName, os.Stderr)
}

type logger struct {
	log *logrus.Logger
}

func newConsoleLogger() *logger {

	log := &logrus.Logger{
		Out:   os.Stdout,
		Level: logrus.InfoLevel,
		Formatter: &easy.Formatter{
			TimestampFormat: "2006-01-02 15:04:05.000",
			LogFormat:       "[%lvl%]: %time% - %msg%\n",
		},
		Hooks: make(logrus.LevelHooks),
	}

	return &logger{log: log}
}

func newFileLogger(logName string, consoleOut io.Writer) *logger {

	logPath := GetLogsDir()
	// f, err := os.OpenFile(filepath.Join(LogPath, logName), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	// if err != nil {
	// 	fmt.Println("Failed to create logfile" + err.Error())
	// 	panic(err)
	// }
	lumberjackLogger := &lumberjack.Logger{
		Filename:   filepath.Join(logPath, logName),
		MaxSize:    2,     // Max size in MB
		MaxBackups: 3,     // Max number of old log files to keep
		MaxAge:     30,    // Max age in days to keep a log file
		Compress:   false, // Compress old log files
	}

	log := &logrus.Logger{
		Out: io.MultiWriter(lumberjackLogger, consoleOut),

		Level: logrus.InfoLevel,
		Formatter: &easy.Formatter{
			TimestampFormat: "2006-01-02 15:04:05.000",
			LogFormat:       "[%lvl%]: %time% - %msg%\n",
		},
		Hooks: make(logrus.LevelHooks),
	}

	return &logger{log: log}
}

func SetLogLevel(level string) {
	log.SetLevel(level)
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

type RemoteHook struct {
	levels  []logrus.Level
	emitter logging.RemoteHookEmitter
	enabled bool
}

func newRemoteHook() *RemoteHook {
	return &RemoteHook{
		levels:  []logrus.Level{logrus.InfoLevel, logrus.ErrorLevel, logrus.WarnLevel, logrus.PanicLevel, logrus.FatalLevel},
		enabled: false,
	}
}

func (h *RemoteHook) Configure(emitter logging.RemoteHookEmitter) {
	h.emitter = emitter
}
func (h *RemoteHook) Levels() []logrus.Level {
	return h.levels
}

func (h *RemoteHook) Enabled(enabled bool) {
	h.enabled = enabled
}

func (h *RemoteHook) Fire(entry *logrus.Entry) error {
	if !h.enabled {
		return nil
	}

	logMessage := logging.LogMessage{
		Level:     entry.Level.String(),
		Message:   entry.Message,
		Timestamp: entry.Time.UnixMilli(),
	}

	bytes, err := json.Marshal(logMessage)
	if err != nil {
		return err
	}
	payload := make(map[string]interface{})
	err = json.Unmarshal(bytes, &payload)
	if err != nil {
		return err
	}

	err = h.emitter.Broadcast("logger", payload)
	if err != nil {
		return err
	}

	return nil
}

func RegisterRemoteLoggerHook(emitter logging.RemoteHookEmitter) {
	log.Infof("Remote hook registered")

	remoteHook.Configure(emitter)
	log.AddHook(remoteHook)
}

func RemoveRemoteLoggerHook() {
	remoteHook.Configure(newRemoteHook().emitter)
	log.log.ReplaceHooks(logrus.LevelHooks{})

	log.Infof("Remote hook unregistered")
}

func EnableRemoteLoggerHook(enabled bool) {
	remoteHook.Enabled(enabled)

	//log.Infof("Remote hook enabled: %v", enabled)
}

func (l *logger) SetLevel(level string) {
	ll, err := logrus.ParseLevel(level)
	if err != nil {
		l.Errorf("undefined level %s", level)
		return
	}

	l.log.SetLevel(ll)
	l.Infof("set loglevel: %s", ll.String())
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

func (l *logger) AddHook(hook logrus.Hook) {
	l.log.AddHook(hook)
}
