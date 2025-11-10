package maxAPI

import (
	"context"
	"fmt"

	maxbot "github.com/max-messenger/max-bot-api-client-go"
)

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

func (b *Bot) sendKeyboard(ctx context.Context, keyboard *maxbot.Keyboard, userID int64, msg string) {
	_, err := b.MaxAPI.Messages.Send(ctx, maxbot.NewMessage().
		SetUser(userID).
		AddKeyboard(keyboard).
		SetText(msg).SetFormat("markdown"))
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
