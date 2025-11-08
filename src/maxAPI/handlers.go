package maxAPI

import (
	"context"
	"io"
	"net/http"
	"os"
	"path/filepath"

	maxbot "github.com/max-messenger/max-bot-api-client-go"
	"github.com/max-messenger/max-bot-api-client-go/schemes"

	"digitalUniversity/database"
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
    attachments := u.Message.Body.Attachments
    if len(attachments) == 0 {
        return
    }

    userID := u.Message.Sender.UserId
    uploadType := b.pendingUploads[userID]
    if uploadType == "" {
        b.logger.Warnf("No pending upload for user %d", userID)
        return
    }

    go func() {
        bgCtx := context.Background()

        for _, att := range attachments {
            fileAtt, ok := att.(*schemes.FileAttachment)
            if !ok {
                continue
            }

            if err := b.downloadAndProcessFile(bgCtx, fileAtt, uploadType); err != nil {
                b.logger.Errorf("Failed to process file %s: %v", fileAtt.Filename, err)
            }
        }
        delete(b.pendingUploads, userID)
    }()
}

func (b *Bot) downloadAndProcessFile(ctx context.Context, fileAtt *schemes.FileAttachment, uploadType string) error {
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
        return nil
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

    importer := database.NewCSVImporter(b.db)
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
        return err
    }

    b.logger.Infof("Successfully imported %s", uploadType)

    if err := os.Remove(filePath); err != nil {
        b.logger.Warnf("Failed to delete file %s: %v", filePath, err)
    }
    b.logger.Debugf("Successfully delete file %s", filePath)

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
