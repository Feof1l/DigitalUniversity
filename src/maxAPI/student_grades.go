package maxAPI

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/max-messenger/max-bot-api-client-go/schemes"

	"digitalUniversity/database"
)

const (
	selectSubjectForGradesMsg = "Выберите предмет для просмотра оценок:"
	noGradesMsg               = "По предмету **%s** оценок пока нет."
	gradesListHeader          = "📊 Оценки по предмету **%s**:\n\n"
	payloadPrefixShowGrades   = "show_grades_"
	gradeEntryFormat          = "`%s %s` — **%d**\n"
	statsFooter               = "\n📈 **Статистика:**\n" +
		"• Всего оценок: **%d**\n" +
		"• Средний балл: **%.2f**"
)

func (b *Bot) handleShowGradesStart(ctx context.Context, userID int64, callbackID string) error {
	studentID, err := b.userRepo.GetUserIDByMaxID(userID)
	if err != nil {
		b.logger.Errorf("Failed to get student ID: %v", err)
		return err
	}

	groupID, err := b.userRepo.GetStudentGroupID(studentID)
	if err != nil {
		b.logger.Errorf("Failed to get student group ID: %v", err)
		return err
	}

	subjects, err := b.gradeRepo.GetSubjectsByStudentGroup(groupID)
	if err != nil {
		b.logger.Errorf("Failed to get subjects for group %d: %v", groupID, err)
		return err
	}

	if len(subjects) == 0 {
		b.answerCallbackWithNotification(ctx, callbackID, "У вашей группы пока нет предметов.")
		return nil
	}

	keyboard := b.MaxAPI.Messages.NewKeyboardBuilder()
	for _, subject := range subjects {
		buttonText := fmt.Sprintf("%s\n", subject.SubjectName)
		payload := fmt.Sprintf("show_grades_subj_%d", subject.SubjectID)
		keyboard.AddRow().AddCallback(buttonText, schemes.DEFAULT, payload)
	}

	keyboard.AddRow().AddCallback(btnBackToMenu, schemes.DEFAULT, payloadBackToMenu)

	messageBody := &schemes.NewMessageBody{
		Text:        selectSubjectForGradesMsg,
		Attachments: []interface{}{schemes.NewInlineKeyboardAttachmentRequest(keyboard.Build())},
	}

	answer := &schemes.CallbackAnswer{Message: messageBody}
	_, err = b.MaxAPI.Messages.AnswerOnCallback(ctx, callbackID, answer)
	return err
}

func (b *Bot) handleShowGradesCallback(ctx context.Context, userID int64, callbackID, payload string) error {
	if !strings.HasPrefix(payload, payloadPrefixShowGrades) {
		return fmt.Errorf("invalid show_grades callback payload: %s", payload)
	}

	parts := strings.Split(payload, "_")
	if len(parts) < 4 {
		return fmt.Errorf("invalid show_grades callback payload format: %s", payload)
	}

	if parts[2] == "subj" {
		return b.handleSubjectSelectedForGrades(ctx, userID, callbackID, payload)
	}

	return fmt.Errorf("unknown show_grades callback type: %s", payload)
}

func (b *Bot) handleSubjectSelectedForGrades(ctx context.Context, userID int64, callbackID, payload string) error {
	var subjectID int64
	fmt.Sscanf(payload, "show_grades_subj_%d", &subjectID)

	studentID, err := b.userRepo.GetUserIDByMaxID(userID)
	if err != nil {
		b.logger.Errorf("Failed to get student ID: %v", err)
		return err
	}

	subjectName, err := b.subjectRepo.GetSubjectName(subjectID)
	if err != nil {
		b.logger.Errorf("Failed to get subject name: %v", err)
		subjectName = "Предмет"
	}

	grades, err := b.gradeRepo.GetGradesByStudentAndSubject(studentID, subjectID)
	if err != nil {
		b.logger.Errorf("Failed to get grades: %v", err)
		return err
	}

	var text string
	if len(grades) == 0 {
		text = fmt.Sprintf(noGradesMsg, subjectName)
	} else {
		text = b.formatGradesList(grades, subjectName)
	}

	keyboard := GetStudentKeyboard(b.MaxAPI)

	messageBody := &schemes.NewMessageBody{
		Text:        text,
		Format:      "markdown",
		Attachments: []interface{}{schemes.NewInlineKeyboardAttachmentRequest(keyboard.Build())},
	}

	answer := &schemes.CallbackAnswer{Message: messageBody}
	_, err = b.MaxAPI.Messages.AnswerOnCallback(ctx, callbackID, answer)
	return err
}

func (b *Bot) formatGradeDateTime(gradeDate time.Time) (dateStr, timeStr string) {
	moscowTime := gradeDate.Add(3 * time.Hour)
	dateStr = moscowTime.Format("02.01.2006")
	timeStr = moscowTime.Format("15:04:05")
	return
}

func (b *Bot) formatGradesList(grades []database.Grade, subjectName string) string {
	if len(grades) == 0 {
		return fmt.Sprintf(noGradesMsg, subjectName)
	}

	var sb strings.Builder
	fmt.Fprintf(&sb, gradesListHeader, subjectName)

	var sum int

	for _, grade := range grades {
		dateStr, timeStr := b.formatGradeDateTime(grade.GradeDate)
		fmt.Fprintf(&sb, gradeEntryFormat, dateStr, timeStr, grade.GradeValue)
		sum += grade.GradeValue
	}

	average := float64(sum) / float64(len(grades))
	fmt.Fprintf(&sb, statsFooter, len(grades), average)

	return sb.String()
}
