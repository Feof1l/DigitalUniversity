package maxAPI

import (
	"github.com/jmoiron/sqlx"
)

type Bot struct {
	//Bot *bot.Bot
	db *sqlx.DB
}
