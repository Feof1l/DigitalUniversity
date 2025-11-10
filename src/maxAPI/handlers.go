package maxAPI

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/max-messenger/max-bot-api-client-go/schemes"
)

const (
	teachersMessage         = "Добро пожаловать, преподаватель! 👨‍🏫\nФункционал для преподавателей находится в разработке."
	studentsMessage         = "Добро пожаловать, студент! 🎓\nФункционал для студентов находится в разработке."
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
	unknownMessageText      = "❓ Я не понимаю это сообщение.\n\nИспользуйте кнопки для взаимодействия с ботом."
	unknownMessageAdmin     = "❓ Я не понимаю это сообщение.\n\nИспользуйте кнопки меню для управления:"
	unknownMessageDefault   = "❓ Я не понимаю это сообщение.\n\nИспользуйте команду /start для начала работы с ботом."
	unknownMessageWithStart = "%s\n\nПожалуйста, используйте команду /start для начала работы."
	nextActionMessage       = "Выберите следующее действие:"
)

func (b *Bot) handleBotStarted(ctx context.Context, u *schemes.BotStartedUpdate) {
	sender := u.User

	if err := b.sendMessage(ctx, sender.UserId, welcomeMsg); err != nil {
		b.logger.Errorf("Failed to send start message: %v", err)
		return
	}

	userRole, err := b.getUserRole(sender.UserId)
	if err != nil {
		b.logger.Errorf("Failed to get role from db: %v", err)
		return
	}

	b.sendKeyboardByRole(ctx, sender.UserId, userRole)
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
			time.Sleep(200 * time.Millisecond)

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
				userRole, _ := b.getUserRole(userID)
				b.sendKeyboardByRole(ctx, userID, userRole)
				return
			}

			b.sendSuccessMessage(ctx, userID, uploadType)
		}()
	}
}

func (b *Bot) handleCallback(ctx context.Context, u *schemes.MessageCallbackUpdate) {
	sender := u.Callback.User
	userID := sender.UserId

	var message string
	switch u.Callback.Payload {
	case "uploadStudents":
		message = sendStudentsFileMessage
		b.pendingUploads[sender.UserId] = "students"
	case "uploadTeachers":
		message = sendTeachersFileMessage
		b.pendingUploads[sender.UserId] = "teachers"
	case "uploadSchedule":
		message = sendScheduleFileMessage
		b.pendingUploads[sender.UserId] = "schedule"
	case "showSchedule":
		currentWeekday := int16(time.Now().Weekday())
		if currentWeekday == 0 {
			currentWeekday = 7
		}
		if err := b.sendScheduleForDay(ctx, userID, currentWeekday); err != nil {
			b.logger.Errorf("Failed to send schedule: %v", err)
		}
		return
	default:
		if strings.HasPrefix(u.Callback.Payload, "sch_day_") {
			var day int16
			fmt.Sscanf(u.Callback.Payload, "sch_day_%d", &day)
			if err := b.sendScheduleForDay(ctx, userID, day); err != nil {
				b.logger.Errorf("Failed to send schedule: %v", err)
			}
			return
		}
		b.logger.Warnf("Unknown callback: %s", u.Callback.Payload)
		return
	}

	if err := b.sendMessage(ctx, sender.UserId, message); err != nil {
		b.logger.Errorf("Failed to send callback response: %v", err)
	}
}

func (b *Bot) handleUnexpectedMessage(ctx context.Context, userID int64) {
	userRole, err := b.getUserRole(userID)
	if err != nil {
		b.logger.Errorf("Failed to get role from db: %v", err)
		b.sendMessage(ctx, userID, unknownMessageText)
		return
	}

	switch userRole {
	case "admin":
		b.sendMessage(ctx, userID, unknownMessageAdmin)
		b.sendKeyboard(ctx, GetAdminKeyboard(b.MaxAPI), userID, adminMsg)
	case "teacher", "student":
		b.sendMessage(ctx, userID, fmt.Sprintf(unknownMessageWithStart, unknownMessageText))
	default:
		b.sendMessage(ctx, userID, unknownMessageDefault)
	}

	delete(b.pendingUploads, userID)
}
