package maxAPI

import (
	"context"
	"digitalUniversity/database"
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
	chooseRoleMessage       = "Добро пожаловать! Выберите свою роль:"
)

func (b *Bot) handleBotStarted(ctx context.Context, u *schemes.BotStartedUpdate) {
	sender := u.User

	userRole, err := b.getUserRole(sender.UserId)
	if err != nil {
		b.logger.Errorf("Failed to get role from db: %v", err)
		b.sendMessage(ctx, sender.UserId, unknownMessageDefault)
		return
	}

	if userRole == "super_user" {
		b.superUser[sender.UserId] = true
		b.sendKeyboard(ctx, GetSuperUserKeyboard(b.MaxAPI), sender.UserId, chooseRoleMessage)
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
		b.logger.Warnf("No pending upload for user %+v %d", b.pendingUploads, userID)
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
	case payload == payloadShowSchedule:
		b.handleShowSchedule(ctx, userID, callbackID)
	case payload == payloadMarkGrade:
		b.handleMarkGrade(ctx, userID, callbackID)
	case payload == payloadMarkAttendance:
		b.handleMarkAttendance(ctx, userID, callbackID)
	case payload == payloadShowScore:
		b.handleShowScore(ctx, userID, callbackID)
	case payload == payloadShowAttendance:
		b.handleShowAttendance(ctx, userID, callbackID)
	case payload == payloadBackToMenu:
		b.handleBackToMenu(ctx, userID, callbackID)
	case payload == payloadAdminRole:
		user := database.User{
			UserMaxID: sender.UserId,
			Name:      sender.Name,
			FirstName: sender.FirstName,
			LastName:  sender.LastName,
			RoleID:    2,
		}
		b.setAndActivateRole(ctx, user, callbackID, payload)
	case payload == payloadTeacherRole:
		user := database.User{
			UserMaxID: sender.UserId,
			Name:      sender.Name,
			FirstName: sender.FirstName,
			LastName:  sender.LastName,
			RoleID:    3,
		}
		b.setAndActivateRole(ctx, user, callbackID, payload)
	case payload == payloadStudentRole:
		user := database.User{
			UserMaxID: sender.UserId,
			Name:      sender.Name,
			FirstName: sender.FirstName,
			LastName:  sender.LastName,
			RoleID:    4,
		}
		b.setAndActivateRole(ctx, user, callbackID, payload)
	case payload == payloadBackToRoleSelection:
		b.answerCallbackWithKeyboard(ctx, callbackID, GetSuperUserKeyboard(b.MaxAPI), chooseRoleMessage)
	case strings.HasPrefix(payload, "sch_day_"):
		b.handleScheduleNavigation(ctx, userID, callbackID, payload)
	case strings.HasPrefix(payload, "grade_"):
		b.handleGradeCallback(ctx, userID, callbackID, payload)
	case strings.HasPrefix(payload, "show_grades_"):
		b.handleShowGradesCallback(ctx, userID, callbackID, payload)
	case strings.HasPrefix(payload, "attend_"):
		b.handleAttendanceCallback(ctx, userID, callbackID, payload)
	case strings.HasPrefix(payload, "show_attend_"):
		b.handleShowAttendanceCallback(ctx, userID, callbackID, payload)
	default:
		b.logger.Warnf("Unknown callback: %s", payload)
	}
}

func (b *Bot) handleMarkAttendance(ctx context.Context, userID int64, callbackID string) {
	if err := b.handleMarkAttendanceStart(ctx, userID, callbackID); err != nil {
		b.logger.Errorf("Failed to start attendance marking: %v", err)
	}
}

func (b *Bot) handleShowAttendance(ctx context.Context, userID int64, callbackID string) {
	if err := b.handleShowAttendanceStart(ctx, userID, callbackID); err != nil {
		b.logger.Errorf("Failed to show attendance: %v", err)
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

	keyboard, menuText := b.getMenuByRole(userRole, userID)
	if keyboard == nil {
		b.logger.Warnf("Unknown role: %s", userRole)
		return fmt.Errorf("unknown role: %s", userRole)
	}

	b.logger.Infof("User %d returned to main menu (role: %s)", userID, userRole)
	return b.answerWithKeyboard(ctx, callbackID, menuText, keyboard)
}

func (b *Bot) getMenuByRole(role string, userID int64) (*maxbot.Keyboard, string) {
	switch role {
	case "admin":
		return GetAdminKeyboard(b.MaxAPI, b.superUser[userID]), mainMenuAdminMsg
	case "teacher":
		return GetTeacherKeyboard(b.MaxAPI, b.superUser[userID]), mainMenuTeacherMsg
	case "student":
		return GetStudentKeyboard(b.MaxAPI, b.superUser[userID]), mainMenuStudentMsg
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

	keyboard, _ := b.getMenuByRole(userRole, userID)
	if keyboard != nil {
		b.sendKeyboard(ctx, keyboard, userID, unknownMessage)
	} else {
		b.sendMessage(ctx, userID, unknownMessageDefault)
	}

	delete(b.pendingUploads, userID)
}

func (b *Bot) setAndActivateRole(ctx context.Context, user database.User, callbackID, payload string) {
	tx, err := b.db.Beginx()
	if err != nil {
		return
	}
	defer tx.Rollback()

	if err := b.userRepo.UpdateUserRole(tx, user.UserMaxID, user.RoleID); err != nil {
		b.logger.Errorf("Failed to set role for user: %v", err)
		b.sendMessage(ctx, user.UserMaxID, "Не удалось сохранить роль")
		return
	}

	err = tx.Commit()
	if err != nil {
		b.logger.Errorf("Failed to update role for user: %v", err)
	}

	//b.sendWelcomeWithKeyboard(ctx, user.UserMaxID, payload)

	var keyboard *maxbot.Keyboard
	var msg string

	switch payload {
	case "admin":
		keyboard = GetAdminKeyboard(b.MaxAPI, b.superUser[user.UserID])
		msg = welcomeAdminMsg
	case "teacher":
		keyboard = GetTeacherKeyboard(b.MaxAPI, b.superUser[user.UserID])
		msg = welcomeTeacherMsg
	case "student":
		keyboard = GetStudentKeyboard(b.MaxAPI, b.superUser[user.UserID])
		msg = welcomeStudentMsg
	}

	// msg := fmt.Sprintf("✅ Роль «%s» установлена!", role)
	b.answerCallbackWithKeyboard(ctx, callbackID, keyboard, msg)
}
