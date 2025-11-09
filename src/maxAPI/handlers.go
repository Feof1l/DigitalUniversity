package maxAPI

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	maxbot "github.com/max-messenger/max-bot-api-client-go"
	"github.com/max-messenger/max-bot-api-client-go/schemes"

	"digitalUniversity/database"
	"digitalUniversity/services"
)

func (b *Bot) handleBotStarted(ctx context.Context, u *schemes.BotStartedUpdate) {
    sender := u.User
    chatID := u.GetChatID()

    _, err := b.MaxAPI.Messages.Send(ctx, maxbot.NewMessage().
        SetUser(sender.UserId).
        SetText(welcomeMsg))
    if err != nil && err.Error() != "" {
        b.logger.Errorf("Failed to send start message: %v", err)
        return
    }

    userRepo := database.NewUserRepository(b.db)
    userRole, err := userRepo.GetUserRole(sender.UserId)
    if err != nil && err.Error() != "" {
        b.logger.Errorf("Failed to get role from db: %v", err)
        return
    }

    switch userRole {
        case "admin":
            b.sendAdminKeyboard(ctx, chatID)
        case "teacher":
            _, err := b.MaxAPI.Messages.Send(ctx, maxbot.NewMessage().
                SetChat(chatID).
                SetText("Добро пожаловать, преподаватель! 👨‍🏫\nФункционал для преподавателей находится в разработке."))
            if err != nil && err.Error() != "" {
                b.logger.Errorf("Failed to send teacher message: %v", err)
            }
        case "student":
            _, err := b.MaxAPI.Messages.Send(ctx, maxbot.NewMessage().
                SetChat(chatID).
                SetText("Добро пожаловать, студент! 🎓\nФункционал для студентов находится в разработке."))
            if err != nil && err.Error() != "" {
                b.logger.Errorf("Failed to send student message: %v", err)
            }
        default:
            b.logger.Warnf("Unknown role for user %d: %q", sender.UserId, userRole)
    }
}

func (b *Bot) sendAdminKeyboard(ctx context.Context, chatID int64) {
    adminKeyboard := GetAdminKeyboard(b.MaxAPI)
    _, err := b.MaxAPI.Messages.Send(ctx, maxbot.NewMessage().
        SetChat(chatID).
        AddKeyboard(adminKeyboard).
        SetText(adminMsg))
    if err != nil && err.Error() != "" {
        b.logger.Errorf("Failed to send admin message: %v", err)
    }
}

func (b *Bot) handleMessageCreated(ctx context.Context, u *schemes.MessageCreatedUpdate) {
    userID := u.Message.Sender.UserId
    messageID := u.Message.Body.Mid

    // ДЕДУПЛИКАЦИЯ по message_id
    b.mu.Lock()
    if b.processedMessages[messageID] {
        b.mu.Unlock()
        b.logger.Debugf("Message %s already processed, skipping", messageID)
        return
    }
    b.processedMessages[messageID] = true
    b.mu.Unlock()

    attachments := u.Message.Body.Attachments
    messageText := u.Message.Body.Text

    if len(attachments) == 0 && messageText != "" {
        b.handleUnexpectedMessage(ctx, userID)

        b.mu.Lock()
        delete(b.processedMessages, messageID)
        b.mu.Unlock()
        return
    }

    if len(attachments) == 0 {
        return
    }

    uploadType := b.pendingUploads[userID]
    if uploadType == "" {
        b.logger.Warnf("No pending upload for user %d", userID)
        return
    }

    fileAttachments := []*schemes.FileAttachment{}
    for _, att := range attachments {
        if fileAtt, ok := att.(*schemes.FileAttachment); ok {
            fileAttachments = append(fileAttachments, fileAtt)
        }
    }

    if len(fileAttachments) == 0 {
        b.sendErrorAndResetUpload(ctx, userID, "Файл не найден. Отправьте CSV файл.")
        return
    }

    if len(fileAttachments) > 1 {
        b.sendErrorAndResetUpload(ctx, userID, fmt.Sprintf("Отправлено %d файла(ов). Пожалуйста, отправьте только один CSV файл за раз.", len(fileAttachments)))
        return
    }

    go func() {
        bgCtx := context.Background()

        if err := b.downloadAndProcessFile(bgCtx, fileAttachments[0], uploadType, userID); err != nil {
            b.logger.Errorf("Failed to process file %s: %v", fileAttachments[0].Filename, err)
            b.sendErrorAndResetUpload(bgCtx, userID, fmt.Sprintf("❌ Ошибка обработки файла:\n\n%v", err))
        }

        delete(b.pendingUploads, userID)

        b.mu.Lock()
        delete(b.processedMessages, messageID)
        b.mu.Unlock()
    }()
}

func (b *Bot) downloadAndProcessFile(ctx context.Context, fileAtt *schemes.FileAttachment, uploadType string, userID int64) error {
    fileURL := fileAtt.Payload.Url
    b.logger.Debugf("Downloading file: %s from %s", fileAtt.Filename, fileURL)

    req, err := http.NewRequestWithContext(ctx, http.MethodGet, fileURL, nil)
    if err != nil {
        return err
    }

    resp, err := http.DefaultClient.Do(req)
    if err != nil {
        return err
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        b.logger.Errorf("Bad HTTP status when downloading file: %s", resp.Status)
        return fmt.Errorf("не удалось скачать файл: статус %s", resp.Status)
    }

    tmpDir := "./tmp"
    if err := os.MkdirAll(tmpDir, 0755); err != nil {
        return err
    }

    filePath := filepath.Join(tmpDir, fileAtt.Filename)

    err = func() error {
        out, err := os.Create(filePath)
        if err != nil {
            return err
        }
        defer out.Close()

        if _, err := io.Copy(out, resp.Body); err != nil {
            return err
        }
        return nil
    }()

    if err != nil {
        b.logger.Errorf("Failed to save file: %v", err)
        return err
    }

    b.logger.Infof("File saved to: %s", filePath)

    var fileType services.FileType
    switch uploadType {
    case "students":
        fileType = services.FileTypeStudents
    case "teachers":
        fileType = services.FileTypeTeachers
    case "schedule":
        fileType = services.FileTypeSchedule
    }

    if err := services.ValidateCSVStructure(filePath, fileType); err != nil {
        os.Remove(filePath)
        return err
    }

    importer := services.NewCSVImporter(b.db)
    switch uploadType {
    case "students":
        err = importer.ImportStudents(filePath)
    case "teachers":
        err = importer.ImportTeachers(filePath)
    case "schedule":
        err = importer.ImportSchedule(filePath)
    default:
        b.logger.Warnf("Unknown upload type: %s", uploadType)
    }

    if err != nil {
        b.logger.Errorf("Failed to import %s: %v", uploadType, err)
        os.Remove(filePath)
        return err
    }

    b.logger.Infof("Successfully imported %s", uploadType)

    b.sendSuccessMessage(ctx, userID, uploadType)

    if err := os.Remove(filePath); err != nil {
        b.logger.Warnf("Failed to delete file %s: %v", filePath, err)
    }
    b.logger.Debugf("Successfully deleted file %s", filePath)

    return nil
}



func (b *Bot) handleCallback(ctx context.Context, u *schemes.MessageCallbackUpdate) {
    sender := u.Callback.User

    var message string
    switch u.Callback.Payload {
    case "uploadStudents":
        message = "Отправьте файл со списком студентов (с расширением .csv)."
        b.pendingUploads[sender.UserId] = "students"
    case "uploadTeachers":
        message = "Отправьте файл с преподавателями (с расширением .csv)."
        b.pendingUploads[sender.UserId] = "teachers"
    case "uploadSchedule":
        message = "Отправьте файл с расписанием (с расширением .csv)."
        b.pendingUploads[sender.UserId] = "schedule"
    default:
        b.logger.Warnf("Unknown callback: %s", u.Callback.Payload)
        return
    }

    _, err := b.MaxAPI.Messages.Send(ctx, maxbot.NewMessage().
        SetUser(sender.UserId).
        SetText(message))
    if err != nil && err.Error() != "" {
        b.logger.Errorf("Failed to send callback response: %v", err)
    }
}

func (b *Bot) sendErrorAndResetUpload(ctx context.Context, userID int64, errorMsg string) {
    _, err := b.MaxAPI.Messages.Send(ctx, maxbot.NewMessage().
        SetUser(userID).
        SetText(fmt.Sprintf("❌ Ошибка:\n\n%s\n\nВыберите действие заново:", errorMsg)))
    if err != nil && err.Error() != "" {
        b.logger.Errorf("Failed to send error message: %v", err)
    }

    adminKeyboard := GetAdminKeyboard(b.MaxAPI)
    _, err = b.MaxAPI.Messages.Send(ctx, maxbot.NewMessage().
        SetUser(userID).
        AddKeyboard(adminKeyboard).
        SetText(adminMsg))
    if err != nil && err.Error() != "" {
        b.logger.Errorf("Failed to send admin keyboard: %v", err)
    }

    delete(b.pendingUploads, userID)
}

func (b *Bot) sendSuccessMessage(ctx context.Context, userID int64, uploadType string) {
    var message string
    switch uploadType {
    case "students":
        message = "✅ Студенты успешно загружены!"
    case "teachers":
        message = "✅ Преподаватели успешно загружены!"
    case "schedule":
        message = "✅ Расписание успешно загружено!"
    }

    _, err := b.MaxAPI.Messages.Send(ctx, maxbot.NewMessage().
        SetUser(userID).
        SetText(message))
    if err != nil && err.Error() != "" {
        b.logger.Errorf("Failed to send success message: %v", err)
    }

    adminKeyboard := GetAdminKeyboard(b.MaxAPI)
    _, err = b.MaxAPI.Messages.Send(ctx, maxbot.NewMessage().
        SetUser(userID).
        AddKeyboard(adminKeyboard).
        SetText("Выберите следующее действие:"))
    if err != nil && err.Error() != "" {
        b.logger.Errorf("Failed to send admin keyboard: %v", err)
    }
}

func (b *Bot) handleUnexpectedMessage(ctx context.Context, userID int64) {
    userRepo := database.NewUserRepository(b.db)
    userRole, err := userRepo.GetUserRole(userID)
    if err != nil && err.Error() != "" {
        b.logger.Errorf("Failed to get role from db: %v", err)
        _, _ = b.MaxAPI.Messages.Send(ctx, maxbot.NewMessage().
            SetUser(userID).
            SetText("❓ Я не понимаю это сообщение.\n\nИспользуйте кнопки для взаимодействия с ботом."))
        return
    }

    switch userRole {
    case "admin":
        _, err := b.MaxAPI.Messages.Send(ctx, maxbot.NewMessage().
            SetUser(userID).
            SetText("❓ Я не понимаю это сообщение.\n\nИспользуйте кнопки меню для управления:"))
        if err != nil && err.Error() != "" {
            b.logger.Errorf("Failed to send unknown message response: %v", err)
        }

        adminKeyboard := GetAdminKeyboard(b.MaxAPI)
        _, err = b.MaxAPI.Messages.Send(ctx, maxbot.NewMessage().
            SetUser(userID).
            AddKeyboard(adminKeyboard).
            SetText(adminMsg))
        if err != nil && err.Error() != "" {
            b.logger.Errorf("Failed to send admin keyboard: %v", err)
        }

    case "teacher":
        _, err := b.MaxAPI.Messages.Send(ctx, maxbot.NewMessage().
            SetUser(userID).
            SetText("❓ Я не понимаю это сообщение.\n\nФункционал для преподавателей находится в разработке.\nПожалуйста, используйте команду /start для начала работы."))
        if err != nil && err.Error() != "" {
            b.logger.Errorf("Failed to send teacher message: %v", err)
        }

    case "student":
        _, err := b.MaxAPI.Messages.Send(ctx, maxbot.NewMessage().
            SetUser(userID).
            SetText("❓ Я не понимаю это сообщение.\n\nФункционал для студентов находится в разработке.\nПожалуйста, используйте команду /start для начала работы."))
        if err != nil && err.Error() != "" {
            b.logger.Errorf("Failed to send student message: %v", err)
        }

    default:
        _, err := b.MaxAPI.Messages.Send(ctx, maxbot.NewMessage().
            SetUser(userID).
            SetText("❓ Я не понимаю это сообщение.\n\nИспользуйте команду /start для начала работы с ботом."))
        if err != nil && err.Error() != "" {
            b.logger.Errorf("Failed to send default message: %v", err)
        }
    }

    delete(b.pendingUploads, userID)
}
