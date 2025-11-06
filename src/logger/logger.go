package logger

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	"digitalUniversity/config"
)


type logMessage struct {
	level  config.LogLevel
	msg    string
	module string
	time   time.Time
}

type Logger struct {
	logDir       string
	level        config.LogLevel
	stdoutLogger *log.Logger
	fileLoggers  map[string]*os.File
	mutex        sync.RWMutex
	msgChan      chan logMessage
	stopChan     chan struct{}
	closed       bool
	wg           sync.WaitGroup
}

var (
	instance *Logger
	once     sync.Once
)

func GetInstance() *Logger {
	once.Do(func() {
		instance = &Logger{
			logDir:      "./logs",
			level:       config.INFO,
			fileLoggers: make(map[string]*os.File),
			msgChan:     make(chan logMessage, 1000),
			stopChan:    make(chan struct{}),
			wg:          sync.WaitGroup{},
		}
		instance.init()
		instance.wg.Add(1)
		go instance.worker()
	})
	return instance
}

func (l *Logger) SetLevel(level config.LogLevel) {
	l.mutex.Lock()
	defer l.mutex.Unlock()
	l.level = level
}

func (l *Logger) SetLogDir(dir string) {
	l.mutex.Lock()
	defer l.mutex.Unlock()
	l.logDir = dir
	_ = os.MkdirAll(l.logDir, 0755)
}

func (l *Logger) init() {
	l.stdoutLogger = log.New(os.Stdout, "", 0)
}

func (l *Logger) getCallerInfo(skip int) (module string) {
	_, file, _, ok := runtime.Caller(skip)
	if !ok {
		return "unknown"
	}
	base := filepath.Base(file)
	return strings.TrimSuffix(base, filepath.Ext(base))
}

func (l *Logger) getFileWriter(moduleName string) *os.File {
	l.mutex.RLock()
	if f, exists := l.fileLoggers[moduleName]; exists {
		l.mutex.RUnlock()
		return f
	}
	l.mutex.RUnlock()

	l.mutex.Lock()
	defer l.mutex.Unlock()
	if f, exists := l.fileLoggers[moduleName]; exists {
		return f
	}

	logPath := filepath.Join(l.logDir, moduleName+".log")
	file, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		fmt.Fprintf(os.Stderr, "[logger] failed to open log file %s: %v\n", logPath, err)
		return nil
	}

	l.fileLoggers[moduleName] = file
	return file
}

func (l *Logger) worker() {
	defer l.wg.Done()
	defer func() {
		l.mutex.Lock()
		for _, f := range l.fileLoggers {
			_ = f.Close()
		}
		l.fileLoggers = make(map[string]*os.File)
		l.mutex.Unlock()
	}()

	for {
		select {
		case msg := <-l.msgChan:
			l.processMessage(msg)
		case <-l.stopChan:
			for len(l.msgChan) > 0 {
				msg := <-l.msgChan
				l.processMessage(msg)
			}
			return
		}
	}
}

func (l *Logger) processMessage(msg logMessage) {
	timestamp := msg.time.Format("2006-01-02 15:04:05")
	formatted := fmt.Sprintf("%s - %s - %s - %s", timestamp, msg.module, msg.level.String(), msg.msg)

	l.stdoutLogger.Println(formatted)

	if file := l.getFileWriter(msg.module); file != nil {
		_, _ = file.WriteString(formatted + "\n")
	}
}

func (l *Logger) log(level config.LogLevel, msg string) {
	l.mutex.RLock()
	if l.closed {
		l.mutex.RUnlock()
		return
	}
	currentLevel := l.level
	l.mutex.RUnlock()

	if level < currentLevel {
		return
	}

	moduleName := l.getCallerInfo(3)

	logMsg := logMessage{
		level:  level,
		msg:    msg,
		module: moduleName,
		time:   time.Now(),
	}

	select {
	case l.msgChan <- logMsg:
	default:
		fmt.Fprintf(os.Stderr, "[logger] dropped log message (channel full): %s\n", msg)
	}
}

func (l *Logger) Debug(msg string)    { l.log(config.DEBUG, msg) }
func (l *Logger) Info(msg string)     { l.log(config.INFO, msg) }
func (l *Logger) Warn(msg string)     { l.log(config.WARNING, msg) }
func (l *Logger) Error(msg string)    { l.log(config.ERROR, msg) }
func (l *Logger) Critical(msg string) { l.log(config.CRITICAL, msg) }

func (l *Logger) Close() {
	l.mutex.Lock()
	if l.closed {
		l.mutex.Unlock()
		return
	}
	l.closed = true
	l.mutex.Unlock()

	close(l.stopChan)
	l.wg.Wait()
}

func (l *Logger) Sync() {
	l.mutex.RLock()
	defer l.mutex.RUnlock()
	for _, f := range l.fileLoggers {
		_ = f.Sync()
	}
}
