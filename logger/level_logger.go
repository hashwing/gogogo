package logger

import (
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"github.com/ivpusic/grpool"
	lumberjack "gopkg.in/natefinch/lumberjack.v2"
)

const (
	LevelError = iota
	LevelWarning
	LevelInformational
	LevelDebug
)

type Logger struct {
	level int
	err   *log.Logger
	warn  *log.Logger
	info  *log.Logger
	debug *log.Logger
	p     *grpool.Pool
	depth int
}

func NewLogger(flag int, numWorkers int, jobQueueLen int, depth int) *Logger {
	Logger := NewLogger3(os.Stdout, flag, numWorkers, jobQueueLen, depth)
	return Logger
}

func NewLogger2(lfn string, maxsize int, flag int, numWorkers int, jobQueueLen int, depth int) *Logger {
	jack := &lumberjack.Logger{
		Filename: lfn,
		MaxSize:  maxsize, // megabytes
	}

	logger := NewLogger3(jack, flag, numWorkers, jobQueueLen, depth)

	return logger
}

func NewLogger3(w io.Writer, flag int, numWorkers int, jobQueueLen int, depth int) *Logger {
	logger := new(Logger)
	logger.depth = depth
	if logger.depth <= 0 {
		logger.depth = 2
	}

	logger.err = log.New(w, "[E] ", flag)
	logger.warn = log.New(w, "[W] ", flag)
	logger.info = log.New(w, "[I] ", flag)
	logger.debug = log.New(w, "[D] ", flag)

	logger.SetLevel(LevelInformational)

	logger.p = grpool.NewPool(numWorkers, jobQueueLen)

	return logger
}

func (ll *Logger) SetLevel(l int) {
	ll.level = l
}

// 统一设置日志前缀
func (ll *Logger) SetPrefix(prefix string) {
	ll.err.SetPrefix(prefix)
	ll.warn.SetPrefix(prefix)
	ll.info.SetPrefix(prefix)
	ll.debug.SetPrefix(prefix)
}

func (ll *Logger) Error(format string, v ...interface{}) {
	if LevelError > ll.level {
		return
	}
	ll.p.JobQueue <- func() {
		ll.err.Output(ll.depth, fmt.Sprintf(format, v...))
	}
}

func (ll *Logger) Warn(format string, v ...interface{}) {
	if LevelWarning > ll.level {
		return
	}
	ll.p.JobQueue <- func() {
		ll.warn.Output(ll.depth, fmt.Sprintf(format, v...))
	}
}

func (ll *Logger) Info(format string, v ...interface{}) {
	if LevelInformational > ll.level {
		return
	}
	ll.p.JobQueue <- func() {
		ll.info.Output(ll.depth, fmt.Sprintf(format, v...))
	}
}

func (ll *Logger) Debug(format string, v ...interface{}) {
	if LevelDebug > ll.level {
		return
	}
	ll.p.JobQueue <- func() {
		ll.debug.Output(ll.depth, fmt.Sprintf(format, v...))
	}
}

func (ll *Logger) SetJack(lfn string, maxsize int) {
	jack := &lumberjack.Logger{
		Filename: lfn,
		MaxSize:  maxsize, // megabytes
	}

	ll.err.SetOutput(jack)
	ll.warn.SetOutput(jack)
	ll.info.SetOutput(jack)
	ll.debug.SetOutput(jack)
}

func (ll *Logger) SetFlag(flag int) {
	ll.err.SetFlags(flag)
	ll.warn.SetFlags(flag)
	ll.debug.SetFlags(flag)
}

func (ll *Logger) Stats() (int, int) {
	return cap(ll.p.JobQueue), len(ll.p.JobQueue)
}

// ================= StdLogger ======================

var (
	StdLogger *Logger = NewLogger(log.LstdFlags|log.Lshortfile, 100, 50, 3)
)

func SetJack(lfn string, maxsize int) {
	StdLogger.SetJack(lfn, maxsize)
}

func Errorf(format string, v ...interface{}) {
	StdLogger.Error(format, v...)
}

func Warnf(format string, v ...interface{}) {
	StdLogger.Warn(format, v...)
}

func Infof(format string, v ...interface{}) {
	StdLogger.Info(format, v...)
}

func Debugf(format string, v ...interface{}) {
	StdLogger.Debug(format, v...)
}

func Error(v ...interface{}) {
	StdLogger.Error(GenerateFmtStr(len(v)), v...)
}

func Warn(v ...interface{}) {
	StdLogger.Warn(GenerateFmtStr(len(v)), v...)
}

func Info(v ...interface{}) {
	StdLogger.Info(GenerateFmtStr(len(v)), v...)
}

func Debug(v ...interface{}) {
	StdLogger.Debug(GenerateFmtStr(len(v)), v...)
}

func LogLevel(logLevel string) string {
	if len(logLevel) == 0 {
		logLevel = "info"
	}
	updateLevel(logLevel)
	Warn("Set Log Level as", logLevel)
	return logLevel
}

func updateLevel(logLevel string) {
	switch strings.ToLower(logLevel) {
	case "debug":
		StdLogger.SetLevel(LevelDebug)
	case "info":
		StdLogger.SetLevel(LevelInformational)
	case "warn":
		StdLogger.SetLevel(LevelWarning)
	case "error":
		StdLogger.SetLevel(LevelError)
	default:
		StdLogger.SetLevel(LevelInformational)
	}
}

func GenerateFmtStr(n int) string {
	return strings.Repeat("%v ", n)
}
