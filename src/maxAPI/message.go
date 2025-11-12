package maxAPI

import (
	"context"
	"fmt"

	maxbot "github.com/max-messenger/max-bot-api-client-go"
	"github.com/max-messenger/max-bot-api-client-go/schemes"
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
	b.sendKeyboardAfterError(ctx, userID)
	delete(b.pendingUploads, userID)
}

func (b *Bot) sendSuccessMessage(ctx context.Context, userID int64, uploadType string) {
	message := b.getSuccessMessage(uploadType)
	b.sendMessage(ctx, userID, message)
	b.sendKeyboard(ctx, GetAdminKeyboard(b.MaxAPI), userID, nextActionMessage)
}

func (b *Bot) getSuccessMessage(uploadType string) string {
	messages := map[string]string{
		"students": studentsSuccessMessage,
		"teachers": teachersSuccessMessage,
		"schedule": scheduleSuccessMessage,
	}

	if msg, exists := messages[uploadType]; exists {
		return msg
	}
	return defaultSuccessMessage
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

func (b *Bot) answerCallbackWithKeyboard(ctx context.Context, callbackID string, keyboard *maxbot.Keyboard, text string) error {
	b.logger.Debugf("Answering callback ID: %s with text length: %d", callbackID, len(text))

	messageBody := &schemes.NewMessageBody{
		Text:        text,
		Format:      "markdown",
		Attachments: []any{},
	}

	keyboardBuilt := keyboard.Build()
	messageBody.Attachments = append(messageBody.Attachments, schemes.NewInlineKeyboardAttachmentRequest(keyboardBuilt))

	answer := &schemes.CallbackAnswer{
		Message: messageBody,
	}

	_, err := b.MaxAPI.Messages.AnswerOnCallback(ctx, callbackID, answer)
	if err != nil && err.Error() != "" {
		b.logger.Errorf("AnswerOnCallback failed: %v", err)
		return err
	}

	b.logger.Infof("Successfully answered callback: %s", callbackID)
	return nil
}

func (b *Bot) sendWelcomeWithKeyboard(ctx context.Context, userID int64, role string) {
	var keyboard *maxbot.Keyboard
	var msg string

	switch role {
	case "admin":
		keyboard = GetAdminKeyboard(b.MaxAPI)
		msg = welcomeAdminMsg
	case "teacher":
		keyboard = GetTeacherKeyboard(b.MaxAPI)
		msg = welcomeTeacherMsg
	case "student":
		keyboard = GetStudentKeyboard(b.MaxAPI)
		msg = welcomeStudentMsg
	default:
		b.logger.Warnf("Unknown role: %q", role)
		b.sendMessage(ctx, userID, "Добро пожаловать! 👋")
		return
	}

	messageID, err := b.MaxAPI.Messages.Send(ctx, maxbot.NewMessage().
		SetUser(userID).
		AddKeyboard(keyboard).
		SetText(msg))
	if err != nil && err.Error() != "" {
		b.logger.Errorf("Failed to send welcome with keyboard: %v", err)
		return
	}

	b.mu.Lock()
	b.lastMessageID[userID] = messageID
	b.mu.Unlock()

	b.logger.Debugf("Saved message ID %s for user %d", messageID, userID)
}

func (b *Bot) sendKeyboardAfterError(ctx context.Context, userID int64) {
	userRole, err := b.getUserRole(userID)
	if err != nil {
		b.logger.Errorf("Failed to get user role: %v", err)
		return
	}

	var keyboard *maxbot.Keyboard
	switch userRole {
	case "admin":
		keyboard = GetAdminKeyboard(b.MaxAPI)
	case "teacher":
		keyboard = GetTeacherKeyboard(b.MaxAPI)
	case "student":
		keyboard = GetStudentKeyboard(b.MaxAPI)
	default:
		b.logger.Warnf("Unknown role for user %d: %q", userID, userRole)
		return
	}

	b.sendKeyboard(ctx, keyboard, userID, retryActionMessage)
}
