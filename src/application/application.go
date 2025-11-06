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
    return &Application{
        logger: logger.GetInstance(),
    }
}

func (app *Application) Configure(cfg *config.Config, ctx context.Context) error {
    db, err := database.OpenDB(&cfg.Database)
    if err != nil {
        app.logger.Errorf("failed open DB: %v", err)
        return err
    }

    app.DB = db

    // bot, err := telegram.NewBot(&cfg.Telegram, app.poller, cfg.Standup.ExcludeList, app.Conversation, db)
    // if err != nil {
    //     app.logger.Errorf("failed create telegram bot: %v", err)
    //     return err
    // }

    // app.Bot = bot

    return nil
}

func (app *Application) Run(ctx context.Context) {
    // go app.Bot.Start(ctx)

    if err := app.DB.Close(); err != nil {
        app.logger.Errorf("failed to close DB: %v", err)
    }
}
