package config

import (
	"log"
	"os"
	"strings"
	"sync"

	"github.com/joho/godotenv"

)

type LogLevel int

const (
	DEBUG LogLevel = iota
	INFO
	WARNING
	ERROR
	CRITICAL
)

type Config struct {
	LogLevel LogLevel
	LogDir   string
}

var instance *Config
var once sync.Once

func Get() *Config {
	once.Do(func() {
		err := godotenv.Load("../.env")
		if err != nil {
			log.Printf("[config] .env not found, using defaults")
		}

		instance = &Config{
			LogLevel: parseLogLevel(os.Getenv("LOG_LEVEL")),
			LogDir:   getEnv("LOG_DIR", "./logs"),
		}
	})
	return instance
}

func parseLogLevel(levelStr string) LogLevel {
	switch strings.ToUpper(strings.TrimSpace(levelStr)) {
	case "DEBUG":
		return DEBUG
	case "INFO":
		return INFO
	case "WARNING":
		return WARNING
	case "ERROR":
		return ERROR
	case "CRITICAL":
		return CRITICAL
	default:
		return INFO
	}
}

func (l LogLevel) String() string {
	switch l {
	case DEBUG:
		return "DEBUG"
	case INFO:
		return "INFO"
	case WARNING:
		return "WARNING"
	case ERROR:
		return "ERROR"
	case CRITICAL:
		return "CRITICAL"
	default:
		return "UNKNOWN"
	}
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
