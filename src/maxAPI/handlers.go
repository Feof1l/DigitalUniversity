package maxAPI

import (
	"context"
	"fmt"
	"strings"
	"time"

	maxbot "github.com/max-messenger/max-bot-api-client-go"
	"github.com/max-messenger/max-bot-api-client-go/schemes"
)

const (
	welcomeTeacherMsg = "Добро пожаловать, преподаватель! 👨‍🏫"
	welcomeStudentMsg = "Добро пожаловать, студент! 🎓"
	welcomeAdminMsg   = "Добро пожаловать, администратор! 👨‍💼"

	mainMenuAdminMsg   = "Главное меню администратора:"
	mainMenuTeacherMsg = "Главное меню преподавателя:"
	mainMenuStudentMsg = "Главное меню студента:"

	unknownMessage        = "❓ Я не понимаю это сообщение."
	unknownMessageDefault = "❓ Я не понимаю это сообщение.\n\nОбратитесь к администратору для получения доступа."
	retryActionMessage    = "Попробуйте снова или выберите другое действие:"

	fileNotFoundMessage     = "Файл не найден. Отправьте CSV файл."
	multipleFilesMessage    = "Отправлено %d файла(ов). Пожалуйста, отправьте только один CSV файл за раз."
	sendStudentsFileMessage = "Отправьте файл со списком студентов (с расширением .csv)."
	sendTeachersFileMessage = "Отправьте файл с преподавателями (с расширением .csv)."
	sendScheduleFileMessage = "Отправьте файл с расписанием (с расширением .csv)."
	errorMessage            = "❌ Ошибка:\n\n%s\n\n"
	studentsSuccessMessage  = "✅ Студенты успешно загружены!"
	teachersSuccessMessage  = "✅ Преподаватели успешно загружены!"
	scheduleSuccessMessage  = "✅ Расписание успешно загружено!"
	defaultSuccessMessage   = "✅ Данные успешно загружены!"
	nextActionMessage       = "Выберите следующее действие:"
)

func (b *Bot) handleBotStarted(ctx context.Context, u *schemes.BotStartedUpdate) {
	sender := u.User

	userRole, err := b.getUserRole(sender.UserId)
	if err != nil {
		b.logger.Errorf("Failed to get role from db: %v", err)
		return
	}

	b.sendWelcomeWithKeyboard(ctx, sender.UserId, userRole)
}

func (b *Bot) handleMessageCreated(ctx context.Context, u *schemes.MessageCreatedUpdate) {
	userID := u.Message.Sender.UserId
	messageID := u.Message.Body.Mid

	if b.isMessageProcessed(messageID) {
		b.logger.Debugf("Message %s already processed, skipping", messageID)
		return
	}

	b.markMessageProcessed(messageID)
	defer b.cleanupProcessedMessage(messageID)

	attachments := u.Message.Body.Attachments
	messageText := u.Message.Body.Text

	if len(attachments) == 0 && messageText != "" {
		b.handleUnexpectedMessage(ctx, userID)
		return
	}

	if len(attachments) == 0 {
		return
	}

	uploadType := b.pendingUploads[userID]
	if uploadType == "" {
		b.logger.Warnf("No pending upload for user %d", userID)
		b.handleUnexpectedMessage(ctx, userID)
		return
	}

	fileAttachments := b.extractFileAttachments(attachments)

	if len(fileAttachments) == 0 {
		b.sendErrorAndResetUpload(ctx, userID, fileNotFoundMessage)
		return
	}

	b.mu.Lock()
	b.uploadCounter[userID]++
	count := b.uploadCounter[userID]
	b.mu.Unlock()

	if count == 1 {
		go func() {
			time.Sleep(500 * time.Millisecond)

			b.mu.Lock()
			totalFiles := b.uploadCounter[userID]
			delete(b.uploadCounter, userID)
			delete(b.pendingUploads, userID)
			b.mu.Unlock()

			if totalFiles > 1 {
				b.sendErrorAndResetUpload(ctx, userID, fmt.Sprintf(multipleFilesMessage, totalFiles))
				return
			}

			if err := b.downloadAndProcessFile(ctx, fileAttachments[0], uploadType); err != nil {
				b.logger.Errorf("Failed to process file %s: %v", fileAttachments[0].Filename, err)
				b.sendMessage(ctx, userID, fmt.Sprintf(errorMessage, err.Error()))
				b.sendKeyboardAfterError(ctx, userID)
				return
			}

			b.sendSuccessMessage(ctx, userID, uploadType)
		}()
	}
}

func (b *Bot) handleCallback(ctx context.Context, u *schemes.MessageCallbackUpdate) {
	sender := u.Callback.User
	userID := sender.UserId
	callbackID := u.Callback.CallbackID
	payload := u.Callback.Payload

	messageID := ""
	if u.Message != nil {
		messageID = u.Message.Body.Mid
	}

	if messageID == "" {
		b.mu.Lock()
		messageID = b.lastMessageID[userID]
		b.mu.Unlock()
	}

	b.logger.Debugf("Callback received: payload=%s, callbackID=%s, userID=%d, messageID=%s",
		payload, callbackID, userID, messageID)

	if msg, uploadType := b.getUploadMessage(payload); msg != "" {
		b.pendingUploads[userID] = uploadType
		if err := b.sendMessage(ctx, userID, msg); err != nil {
			b.logger.Errorf("Failed to send callback response: %v", err)
		}
		return
	}

	switch {
	case payload == "showSchedule":
		b.handleShowSchedule(ctx, userID, callbackID)
	case payload == "markGrade":
		b.handleMarkGrade(ctx, userID, callbackID)
	case payload == "showScore":
		b.handleShowScore(ctx, userID, callbackID)
	case payload == "backToMenu":
		b.handleBackToMenu(ctx, userID, callbackID)
	case strings.HasPrefix(payload, "sch_day_"):
		b.handleScheduleNavigation(ctx, userID, callbackID, payload)
	case strings.HasPrefix(payload, "grade_"):
		b.handleGradeCallback(ctx, userID, callbackID, payload)
	case strings.HasPrefix(payload, "show_grades_"):
		b.handleShowGradesCallback(ctx, userID, callbackID, payload)
	default:
		b.logger.Warnf("Unknown callback: %s", payload)
	}
}

func (b *Bot) getUploadMessage(payload string) (string, string) {
	switch payload {
	case "uploadStudents":
		return sendStudentsFileMessage, "students"
	case "uploadTeachers":
		return sendTeachersFileMessage, "teachers"
	case "uploadSchedule":
		return sendScheduleFileMessage, "schedule"
	default:
		return "", ""
	}
}

func (b *Bot) handleShowSchedule(ctx context.Context, userID int64, callbackID string) {
	currentWeekday := int16(time.Now().Weekday())
	if currentWeekday == 0 {
		currentWeekday = 7
	}
	if err := b.sendScheduleForDay(ctx, userID, callbackID, currentWeekday); err != nil {
		b.logger.Errorf("Failed to send schedule: %v", err)
	}
}

func (b *Bot) handleMarkGrade(ctx context.Context, userID int64, callbackID string) {
	if err := b.handleMarkGradeStart(ctx, userID, callbackID); err != nil {
		b.logger.Errorf("Failed to start grade marking: %v", err)
	}
}

func (b *Bot) handleShowScore(ctx context.Context, userID int64, callbackID string) {
	if err := b.handleShowGradesStart(ctx, userID, callbackID); err != nil {
		b.logger.Errorf("Failed to show grades: %v", err)
	}
}

func (b *Bot) handleScheduleNavigation(ctx context.Context, userID int64, callbackID, payload string) {
	var day int16
	fmt.Sscanf(payload, "sch_day_%d", &day)

	b.logger.Debugf("Processing schedule navigation: day=%d, callbackID=%s", day, callbackID)

	if err := b.answerScheduleCallback(ctx, userID, callbackID, day); err != nil {
		b.logger.Errorf("Failed to answer callback: %v", err)
	}
}

func (b *Bot) handleBackToMenu(ctx context.Context, userID int64, callbackID string) error {
	userRole, err := b.getUserRole(userID)
	if err != nil {
		b.logger.Errorf("Failed to get user role: %v", err)
		return err
	}

	keyboard, menuText := b.getMenuByRole(userRole)
	if keyboard == nil {
		b.logger.Warnf("Unknown role: %s", userRole)
		return fmt.Errorf("unknown role: %s", userRole)
	}

	messageBody := &schemes.NewMessageBody{
		Text:        menuText,
		Attachments: []interface{}{schemes.NewInlineKeyboardAttachmentRequest(keyboard.Build())},
	}

	answer := &schemes.CallbackAnswer{Message: messageBody}
	_, err = b.MaxAPI.Messages.AnswerOnCallback(ctx, callbackID, answer)
	if err != nil && err.Error() != "" {
		b.logger.Errorf("Failed to answer callback: %v", err)
		return err
	}

	b.logger.Infof("User %d returned to main menu (role: %s)", userID, userRole)
	return nil
}

func (b *Bot) getMenuByRole(role string) (*maxbot.Keyboard, string) {
	switch role {
	case "admin":
		return GetAdminKeyboard(b.MaxAPI), mainMenuAdminMsg
	case "teacher":
		return GetTeacherKeyboard(b.MaxAPI), mainMenuTeacherMsg
	case "student":
		return GetStudentKeyboard(b.MaxAPI), mainMenuStudentMsg
	default:
		return nil, ""
	}
}

func (b *Bot) handleUnexpectedMessage(ctx context.Context, userID int64) {
	userRole, err := b.getUserRole(userID)
	if err != nil {
		b.logger.Errorf("Failed to get role from db: %v", err)
		b.sendMessage(ctx, userID, unknownMessageDefault)
		return
	}

	keyboard, _ := b.getMenuByRole(userRole)
	if keyboard != nil {
		b.sendKeyboard(ctx, keyboard, userID, unknownMessage)
	} else {
		b.sendMessage(ctx, userID, unknownMessageDefault)
	}

	delete(b.pendingUploads, userID)
}
