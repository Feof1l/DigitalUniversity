package application

import (
	"context"

	"github.com/jmoiron/sqlx"

	"digitalUniversity/config"
	"digitalUniversity/database"
	"digitalUniversity/logger"
	"digitalUniversity/maxAPI"
)

type Application struct {
	Bot    *maxAPI.Bot
	DB     *sqlx.DB
	logger *logger.Logger
}

func NewApplication() *Application {
	return &Application{}
}

func (app *Application) Configure(cfg *config.Config, logger *logger.Logger, ctx context.Context) error {
	app.logger = logger

	db, err := database.OpenDB(&cfg.Database)
	if err != nil {
		return err
	}

	app.DB = db

	b, err := maxAPI.NewBot(&cfg.MaxAPI, logger, db, ctx)
	if err != nil {
		_ = db.Close()
		return err
	}
	app.Bot = b

	return nil
}

func (app *Application) Run(ctx context.Context) {
	app.Bot.Start(ctx)
}
