package maxAPI

import (
	"context"
	"digitalUniversity/config"
	"digitalUniversity/database"
	"digitalUniversity/logger"

	"github.com/jmoiron/sqlx"
	maxbot "github.com/max-messenger/max-bot-api-client-go"
	"github.com/max-messenger/max-bot-api-client-go/schemes"
)

type Bot struct {
	MaxBot *schemes.BotInfo
	db     *sqlx.DB
	logger *logger.Logger
	MaxAPI *maxbot.Api
}

const (
	welcomeMsg = "Привет! 👋 Я бот цифрового университета."
	adminMsg   = "Вы администратор. Выберите действие:"
)

type Role string

const (
	ADMIN   Role = "Администратор"
	STUDENT Role = "Студент"
	TEACHER Role = "Преподаватель"
)

func NewMaxBot(t *Bot, config *config.MaxConfig, ctx context.Context) (*maxbot.Api, *schemes.BotInfo, error) {
	api, err := maxbot.New(config.Token)
	if err != nil {
		return nil, nil, err
	}

	b, err := api.Bots.GetBot(ctx)
	if err != nil {
		return nil, nil, err
	}

	return api, b, nil
}

func NewBot(config *config.MaxConfig, logger *logger.Logger, db *sqlx.DB, ctx context.Context) (*Bot, error) {
	b := &Bot{
		db:     db,
		logger: logger,
	}

	api, maxBot, err := NewMaxBot(b, config, ctx)
	if err != nil {
		b.logger.Errorf("failed create telegram bot %v", err)
		return nil, err
	}

	b.MaxBot = maxBot
	b.MaxAPI = api

	return b, nil
}

func (b *Bot) Start(ctx context.Context) {
	go func() {
		for upd := range b.MaxAPI.GetUpdates(ctx) {
			//b.MaxAPI.Debugs.Send(ctx, upd)
			b.logger.Infof("Received update: %#v", upd)

			switch u := upd.(type) {
			case *schemes.BotStartedUpdate:
				sender := u.User
				chatID := u.GetChatID()

				_, err := b.MaxAPI.Messages.Send(ctx, maxbot.NewMessage().
					SetUser(sender.UserId).
					SetText(welcomeMsg))
				if err != nil {
					b.logger.Errorf("Failed to send start message %v", err)
				}

				userRole, err := database.GetUserRole(b.db, sender.UserId)
				if err != nil {
					b.logger.Errorf("Failed to get role from db %v", err)
				}

				if userRole == string(ADMIN) {
					adminKeyboard := GetKeyboard(b.MaxAPI, ctx)

					_, err := b.MaxAPI.Messages.Send(ctx, maxbot.NewMessage().SetChat(chatID).AddKeyboard(adminKeyboard).SetText(adminMsg))
					if err != nil {
						b.logger.Errorf("Failed to send msg %v", err)
					}

				}

			case *schemes.MessageCreatedUpdate:
				attachments := u.Message.Body.Attachments

				if len(attachments) == 0 {
					continue
				}

				_, err := b.MaxAPI.Uploads.UploadMediaFromFile(ctx, schemes.FILE, "./file.csv")
				if err != nil {
					b.logger.Errorf("Failed to upload file %v", err)
				}

			case *schemes.MessageCallbackUpdate:
				sender := u.Callback.User

				switch u.Callback.Payload {
				case "uploadStudents":
					_, err := b.MaxAPI.Messages.Send(ctx, maxbot.NewMessage().
						SetUser(sender.UserId).
						SetText("Отправьте файл со списком стдунетов (с расширением .csv)."))
					if err != nil {
						b.logger.Errorf("Failed to request students file: %v", err)
					}

				case "uploadTeachers":
					_, err := b.MaxAPI.Messages.Send(ctx, maxbot.NewMessage().
						SetUser(sender.UserId).
						SetText("Отправьте файл с преподавателями (с расширением .csv)."))
					if err != nil {
						b.logger.Errorf("Failed to request teachers file: %v", err)
					}
				case "uploadSchedule":
					_, err := b.MaxAPI.Messages.Send(ctx, maxbot.NewMessage().
						SetUser(sender.UserId).
						SetText("Отправьте файл с расписанием (с расширением .csv)."))
					if err != nil {
						b.logger.Errorf("Failed to request schedule file: %v", err)
					}
				}

			default:
				b.logger.Debugf("Unhandled update type: %T", upd)
			}
		}
	}()
}
