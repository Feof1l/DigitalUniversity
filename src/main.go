package main

import (
	"digitalUniversity/config"
	"digitalUniversity/logger"
)

func main() {
	cfg := config.Get()
	log := logger.GetInstance()
	defer log.Close()

	log.SetLevel(cfg.LogLevel)
	log.SetLogDir(cfg.LogDir)

	log.Info("Application started with config: LOG_LEVEL=" + cfg.LogLevel.String())
	log.Warn("This is a warning")
	log.Error("Something went wrong")
}
