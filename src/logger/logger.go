package logger

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"sync"

	"gopkg.in/natefinch/lumberjack.v2"
)

type LogLevel int

const (
    DEBUG LogLevel = iota
    INFO
    WARNING
    ERROR
    CRITICAL
)

var levelNames = map[LogLevel]string{
    DEBUG:    "DEBUG",
    INFO:     "INFO",
    WARNING:  "WARNING",
    ERROR:    "ERROR",
    CRITICAL: "CRITICAL",
}

type Logger struct {
    level LogLevel
    out   *log.Logger
    mu    sync.RWMutex
}

var instance *Logger
var once sync.Once

func GetInstance() *Logger {
    once.Do(func() {
        instance = &Logger{
            level: CRITICAL,
            out:   log.New(os.Stderr, "", log.LstdFlags|log.Lshortfile),
        }
    })
    return instance
}

func (l *Logger) Initialize(logDir string, level LogLevel) error {
    l.mu.Lock()
    defer l.mu.Unlock()

    l.level = level

    if err := os.MkdirAll(logDir, 0755); err != nil {
        return fmt.Errorf("failed to create log directory: %w", err)
    }

    logFile := filepath.Join(logDir, "app.log")

    fileWriter := &lumberjack.Logger{
        Filename:   logFile,
        MaxSize:    100,
        MaxBackups: 3,
        MaxAge:     28,
        Compress:   true,
    }

    multiWriter := io.MultiWriter(os.Stdout, fileWriter)
    l.out = log.New(multiWriter, "", log.LstdFlags|log.Lshortfile)

    return nil
}

func (l *Logger) SetLevel(level LogLevel) {
    l.mu.Lock()
    defer l.mu.Unlock()
    l.level = level
}

func (l *Logger) log(level LogLevel, msg string) {
    l.mu.RLock()
    currentLevel := l.level
    l.mu.RUnlock()

    if level < currentLevel {
        return
    }

    l.out.Output(3, fmt.Sprintf("[%s] %s", levelNames[level], msg))
}

func (l *Logger) Debug(msg string)    { l.log(DEBUG, msg) }
func (l *Logger) Info(msg string)     { l.log(INFO, msg) }
func (l *Logger) Warn(msg string)     { l.log(WARNING, msg) }
func (l *Logger) Error(msg string)    { l.log(ERROR, msg) }
func (l *Logger) Critical(msg string) { l.log(CRITICAL, msg) }

func (l *Logger) Debugf(format string, args ...interface{}) {
    l.log(DEBUG, fmt.Sprintf(format, args...))
}
func (l *Logger) Infof(format string, args ...interface{}) {
    l.log(INFO, fmt.Sprintf(format, args...))
}
func (l *Logger) Warnf(format string, args ...interface{}) {
    l.log(WARNING, fmt.Sprintf(format, args...))
}
func (l *Logger) Errorf(format string, args ...interface{}) {
    l.log(ERROR, fmt.Sprintf(format, args...))
}
func (l *Logger) Criticalf(format string, args ...interface{}) {
    l.log(CRITICAL, fmt.Sprintf(format, args...))
}

func (l *Logger) Fatal(msg string) {
    l.log(CRITICAL, msg)
    os.Exit(1)
}

func (l *Logger) Fatalf(format string, args ...interface{}) {
    l.log(CRITICAL, fmt.Sprintf(format, args...))
    os.Exit(1)
}
