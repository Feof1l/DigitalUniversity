package maxAPI

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	maxbot "github.com/max-messenger/max-bot-api-client-go"
	"github.com/max-messenger/max-bot-api-client-go/schemes"

	"digitalUniversity/services"
)

const (
	ErrDownloadStatusFmt = "failed to download file: status %s"

	UnknownUploadTypeWarnFmt = "Unknown upload type: %s"
	UnknownUploadTypeErrFmt  = "unknown upload type: %s"

	UnknownSubjectText = "Неизвестный предмет"
	UnknownLessonType  = "Неизвестный тип"
	UnknownTeacher     = "Неизвестный преподаватель"
	UnknownGroup       = "Неизвестная группа"
)

func (b *Bot) downloadFile(ctx context.Context, fileAtt *schemes.FileAttachment) (string, error) {
	fileURL := fileAtt.Payload.Url
	b.logger.Debugf("Downloading file: %s from %s", fileAtt.Filename, fileURL)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fileURL, nil)
	if err != nil {
		return "", err
	}

	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		b.logger.Errorf("Bad HTTP status when downloading file: %s", resp.Status)
		return "", fmt.Errorf(ErrDownloadStatusFmt, resp.Status)
	}

	tmpDir := "./tmp"
	if err := os.MkdirAll(tmpDir, 0o755); err != nil {
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

	records, err := services.ValidateCSVStructure(filePath, fileType)
	if err != nil {
		return err
	}

	importer := services.NewCSVImporter(b.db)
	switch uploadType {
	case "students":
		return importer.ImportStudents(records)
	case "teachers":
		return importer.ImportTeachers(records)
	case "schedule":
		return importer.ImportSchedule(records)
	default:
		b.logger.Warnf(UnknownUploadTypeWarnFmt, uploadType)
		return fmt.Errorf(UnknownUploadTypeErrFmt, uploadType)
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

func (b *Bot) getUserRole(userID int64) (string, error) {
	return b.userRepo.GetUserRole(userID)
}

func (b *Bot) extractFileAttachments(attachments []any) []*schemes.FileAttachment {
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

func (b *Bot) getNearestDateForWeekday(targetWeekday int16) time.Time {
	now := time.Now().Local()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	goWeekday := time.Weekday(targetWeekday % 7)

	daysAhead := int(goWeekday - today.Weekday())
	if daysAhead < 0 {
		daysAhead += 7
	}

	return today.AddDate(0, 0, daysAhead)
}

func (b *Bot) getSubjectName(subjectID int64) string {
	name, err := b.subjectRepo.GetSubjectName(subjectID)
	if err != nil {
		b.logger.Errorf("Failed to get subject name for ID %d: %v", subjectID, err)
		return UnknownSubjectText
	}
	return name
}

func (b *Bot) getLessonTypeName(lessonTypeID int64) string {
	name, err := b.lessonTypeRepo.GetLessonTypeName(lessonTypeID)
	if err != nil {
		b.logger.Errorf("Failed to get lesson type name for ID %d: %v", lessonTypeID, err)
		return UnknownLessonType
	}
	return name
}

func (b *Bot) getTeacherName(teacherID int64) string {
	name, err := b.userRepo.GetTeacherName(teacherID)
	if err != nil {
		b.logger.Errorf("Failed to get teacher name for ID %d: %v", teacherID, err)
		return UnknownTeacher
	}
	return name
}

func (b *Bot) getGroupName(groupID int64) string {
	name, err := b.groupRepo.GetGroupName(groupID)
	if err != nil {
		b.logger.Errorf("Failed to get group name for ID %d: %v", groupID, err)
		return UnknownGroup
	}
	return name
}

func (b *Bot) answerWithKeyboard(ctx context.Context, callbackID string, text string, keyboard *maxbot.Keyboard) error {
	messageBody := &schemes.NewMessageBody{
		Text:        text,
		Attachments: []any{schemes.NewInlineKeyboardAttachmentRequest(keyboard.Build())},
	}
	answer := &schemes.CallbackAnswer{Message: messageBody}
	_, err := b.MaxAPI.Messages.AnswerOnCallback(ctx, callbackID, answer)
	return err
}

func (b *Bot) answerWithKeyboardAndNotification(ctx context.Context, callbackID string, text string, keyboard *maxbot.Keyboard, notification string) error {
	messageBody := &schemes.NewMessageBody{
		Text:        text,
		Attachments: []any{schemes.NewInlineKeyboardAttachmentRequest(keyboard.Build())},
	}
	answer := &schemes.CallbackAnswer{
		Message:      messageBody,
		Notification: notification,
	}
	_, err := b.MaxAPI.Messages.AnswerOnCallback(ctx, callbackID, answer)
	return err
}

func (b *Bot) answerWithKeyboardMarkdown(ctx context.Context, callbackID string, text string, keyboard *maxbot.Keyboard) error {
	messageBody := &schemes.NewMessageBody{
		Text:        text,
		Format:      "markdown",
		Attachments: []any{schemes.NewInlineKeyboardAttachmentRequest(keyboard.Build())},
	}
	answer := &schemes.CallbackAnswer{Message: messageBody}
	_, err := b.MaxAPI.Messages.AnswerOnCallback(ctx, callbackID, answer)
	return err
}
