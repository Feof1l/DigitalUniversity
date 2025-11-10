package maxAPI

import (
	"context"
	"digitalUniversity/services"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	maxbot "github.com/max-messenger/max-bot-api-client-go"
	"github.com/max-messenger/max-bot-api-client-go/schemes"
)

func (b *Bot) sendKeyboard(ctx context.Context, keyboard *maxbot.Keyboard, userID int64, msg string) {
	_, err := b.MaxAPI.Messages.Send(ctx, maxbot.NewMessage().
		SetUser(userID).
		AddKeyboard(keyboard).
		SetText(msg))
	if err != nil && err.Error() != "" {
		b.logger.Errorf("Failed to send keyboard: %v", err)
	}
}

func (b *Bot) sendMessage(ctx context.Context, userID int64, text string) error {
	_, err := b.MaxAPI.Messages.Send(ctx, maxbot.NewMessage().
		SetUser(userID).
		SetText(text))
	if err != nil && err.Error() != "" {
		return err
	}
	return nil
}

func (b *Bot) sendKeyboardByRole(ctx context.Context, userID int64, role string) {
	var keyboard *maxbot.Keyboard
	var msg string

	switch role {
	case "admin":
		keyboard = GetAdminKeyboard(b.MaxAPI)
		msg = adminMsg
	case "teacher":
		keyboard = GetTeacherKeyboard(b.MaxAPI)
		msg = teachersMessage
	case "student":
		keyboard = GetStudentKeyboard(b.MaxAPI)
		msg = studentsMessage
	default:
		b.logger.Warnf("Unknown role: %q", role)
		return
	}

	_, err := b.MaxAPI.Messages.Send(ctx, maxbot.NewMessage().
		SetUser(userID).
		AddKeyboard(keyboard).
		SetText(msg))
	if err != nil && err.Error() != "" {
		b.logger.Errorf("Failed to send keyboard: %v", err)
	}
}

func (b *Bot) downloadFile(ctx context.Context, fileAtt *schemes.FileAttachment) (string, error) {
	fileURL := fileAtt.Payload.Url
	b.logger.Debugf("Downloading file: %s from %s", fileAtt.Filename, fileURL)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fileURL, nil)
	if err != nil {
		return "", err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		b.logger.Errorf("Bad HTTP status when downloading file: %s", resp.Status)
		return "", fmt.Errorf("failed to download file: status %s", resp.Status)
	}

	tmpDir := "./tmp"
	if err := os.MkdirAll(tmpDir, 0755); err != nil {
		return "", err
	}

	filePath := filepath.Join(tmpDir, fileAtt.Filename)

	if err := b.saveFile(filePath, resp.Body); err != nil {
		return "", err
	}

	b.logger.Infof("File saved to: %s", filePath)
	return filePath, nil
}

func (b *Bot) saveFile(filePath string, reader io.Reader) error {
	out, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, reader)
	return err
}

func (b *Bot) validateAndImportFile(filePath, uploadType string) error {
	fileType := b.getFileType(uploadType)

	if err := services.ValidateCSVStructure(filePath, fileType); err != nil {
		return err
	}

	importer := services.NewCSVImporter(b.db)
	switch uploadType {
	case "students":
		return importer.ImportStudents(filePath)
	case "teachers":
		return importer.ImportTeachers(filePath)
	case "schedule":
		return importer.ImportSchedule(filePath)
	default:
		b.logger.Warnf("Unknown upload type: %s", uploadType)
		return fmt.Errorf("unknown upload type: %s", uploadType)
	}
}

func (b *Bot) getFileType(uploadType string) services.FileType {
	switch uploadType {
	case "students":
		return services.FileTypeStudents
	case "teachers":
		return services.FileTypeTeachers
	case "schedule":
		return services.FileTypeSchedule
	default:
		return ""
	}
}

func (b *Bot) sendErrorAndResetUpload(ctx context.Context, userID int64, errorMsg string) {
	b.sendMessage(ctx, userID, fmt.Sprintf(errorMessage, errorMsg))

	userRole, err := b.getUserRole(userID)
	if err != nil {
		b.logger.Errorf("Failed to get user role: %v", err)
		return
	}

	b.sendKeyboardByRole(ctx, userID, userRole)
	delete(b.pendingUploads, userID)
}

func (b *Bot) sendSuccessMessage(ctx context.Context, userID int64, uploadType string) {
	message := b.getSuccessMessage(uploadType)
	b.sendMessage(ctx, userID, message)
	b.sendKeyboard(ctx, GetAdminKeyboard(b.MaxAPI), userID, nextActionMessage)
}

func (b *Bot) getSuccessMessage(uploadType string) string {
	switch uploadType {
	case "students":
		return studentsSuccessMessage
	case "teachers":
		return teachersSuccessMessage
	case "schedule":
		return scheduleSuccessMessage
	default:
		return defaultSuccessMessage
	}
}

func (b *Bot) getUserRole(userID int64) (string, error) {
	return b.userRepo.GetUserRole(userID)
}

func (b *Bot) isMessageProcessed(messageID string) bool {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.processedMessages[messageID]
}

func (b *Bot) markMessageProcessed(messageID string) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.processedMessages[messageID] = true
}

func (b *Bot) cleanupProcessedMessage(messageID string) {
	b.mu.Lock()
	defer b.mu.Unlock()
	delete(b.processedMessages, messageID)
}

func (b *Bot) extractFileAttachments(attachments []interface{}) []*schemes.FileAttachment {
	fileAttachments := []*schemes.FileAttachment{}
	for _, att := range attachments {
		if fileAtt, ok := att.(*schemes.FileAttachment); ok {
			fileAttachments = append(fileAttachments, fileAtt)
		}
	}
	return fileAttachments
}

func (b *Bot) downloadAndProcessFile(ctx context.Context, fileAtt *schemes.FileAttachment, uploadType string) error {
	filePath, err := b.downloadFile(ctx, fileAtt)
	if err != nil {
		return err
	}
	defer os.Remove(filePath)

	if err := b.validateAndImportFile(filePath, uploadType); err != nil {
		return err
	}

	return nil
}
