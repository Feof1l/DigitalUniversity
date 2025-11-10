package maxAPI

import (
	"context"
	"sync"

	"github.com/jmoiron/sqlx"
	maxbot "github.com/max-messenger/max-bot-api-client-go"
	"github.com/max-messenger/max-bot-api-client-go/schemes"

	"digitalUniversity/config"
	"digitalUniversity/database"
	"digitalUniversity/logger"
)

const (
	welcomeMsg = "Привет! 👋 Я бот цифрового университета."
	adminMsg   = "Вы администратор. Выберите действие:"
)

type Bot struct {
	MaxBot            *schemes.BotInfo
	db                *sqlx.DB
	logger            *logger.Logger
	MaxAPI            *maxbot.Api
	pendingUploads    map[int64]string
	processedMessages map[string]bool
	uploadCounter     map[int64]int
	mu                sync.Mutex

	userRepo       *database.UserRepository
	groupRepo      *database.GroupRepository
	subjectRepo    *database.SubjectRepository
	lessonTypeRepo *database.LessonTypeRepository
	scheduleRepo   *database.ScheduleRepository
	attendanceRepo *database.AttendanceRepository

	scheduleMessageIDs map[int64]string
}

func NewBot(cfg *config.MaxConfig, log *logger.Logger, db *sqlx.DB, ctx context.Context) (*Bot, error) {
	api, err := maxbot.New(cfg.Token)
	if err != nil && err.Error() != "" {
		log.Errorf("failed to create max api: %v", err)
		return nil, err
	}

	maxBot, err := api.Bots.GetBot(ctx)
	if err != nil && err.Error() != "" {
		log.Errorf("failed to get bot info: %v", err)
		return nil, err
	}

	return &Bot{
		MaxBot:            maxBot,
		db:                db,
		logger:            log,
		MaxAPI:            api,
		pendingUploads:    make(map[int64]string),
		processedMessages: make(map[string]bool),
		uploadCounter:     make(map[int64]int),

		userRepo:           database.NewUserRepository(db),
		groupRepo:          database.NewGroupRepository(db),
		subjectRepo:        database.NewSubjectRepository(db),
		lessonTypeRepo:     database.NewLessonTypeRepository(db),
		scheduleRepo:       database.NewScheduleRepository(db),
		scheduleMessageIDs: make(map[int64]string),
	}, nil
}

func (b *Bot) Start(ctx context.Context) {
	go func() {
		for upd := range b.MaxAPI.GetUpdates(ctx) {
			b.logger.Debugf("Received update type: %T", upd)

			switch u := upd.(type) {
			case *schemes.BotStartedUpdate:
				b.handleBotStarted(ctx, u)
			case *schemes.MessageCreatedUpdate:
				b.handleMessageCreated(ctx, u)
			case *schemes.MessageCallbackUpdate:
				b.handleCallback(ctx, u)
			default:
				b.logger.Debugf("Unhandled update type: %T", upd)
			}
		}
	}()
}
