package maxAPI

import (
	"context"
	"fmt"
	"strings"

	"github.com/max-messenger/max-bot-api-client-go/schemes"

	"digitalUniversity/database"
)

const (
	scheduleTemplate = "%d. **%s** (%s)\n   👨‍🏫 %s\n   👥 %s\n   🏫 %s\n   ⏰ %s–%s\n\n"
	timeFormat       = "15:04"
)

var weekdayNames = map[int16]string{
	1: "Понедельник",
	2: "Вторник",
	3: "Среда",
	4: "Четверг",
	5: "Пятница",
	6: "Суббота",
	7: "Воскресенье",
}

func (b *Bot) formatSchedule(entries []database.Schedule, weekday int16) string {
	dayName := b.getWeekdayName(weekday)
	date := b.getNearestDateForWeekday(weekday).Format("02.01")

	if len(entries) == 0 {
		return fmt.Sprintf("🗓️%s **%s**\n\nНет занятий.", date, dayName)
	}

	var sb strings.Builder
	fmt.Fprintf(&sb, "🗓️%s **%s**\n\n", date, dayName)

	for i, entry := range entries {
		b.appendScheduleEntry(&sb, i+1, entry)
	}

	return strings.TrimSpace(sb.String())
}

func (b *Bot) getWeekdayName(weekday int16) string {
	if name, exists := weekdayNames[weekday]; exists {
		return name
	}
	return fmt.Sprintf("День %d", weekday)
}

func (b *Bot) appendScheduleEntry(sb *strings.Builder, index int, entry database.Schedule) {
	subjectName := b.getSubjectName(entry.SubjectID)
	lessonTypeName := b.getLessonTypeName(entry.LessonTypeID)
	teacherName := b.getTeacherName(entry.TeacherID)
	groupName := b.getGroupName(entry.GroupID)

	startTime := entry.StartTime.Format(timeFormat)
	endTime := entry.EndTime.Format(timeFormat)

	fmt.Fprintf(sb, scheduleTemplate,
		index,
		subjectName,
		lessonTypeName,
		teacherName,
		groupName,
		entry.ClassRoom,
		startTime,
		endTime,
	)
}

func (b *Bot) calculateNavigationDays(weekday int16) (prev, next int16) {
	prev = weekday - 1
	if prev < 1 {
		prev = 7
	}

	next = weekday + 1
	if next > 7 {
		next = 1
	}

	return prev, next
}

func (b *Bot) getScheduleEntriesForUser(maxUserID int64, weekday int16) ([]database.Schedule, error) {
	userRole, err := b.getUserRole(maxUserID)
	if err != nil {
		b.logger.Errorf("Failed to get user role: %v", err)
		return nil, err
	}

	switch userRole {
	case "teacher":
		teacherID, err := b.userRepo.GetUserIDByMaxID(maxUserID)
		if err != nil {
			b.logger.Errorf("Failed to get teacher ID: %v", err)
			return nil, err
		}
		return b.scheduleRepo.GetScheduleForDateByTeacher(weekday, teacherID)

	case "student":
		studentID, err := b.userRepo.GetUserIDByMaxID(maxUserID)
		if err != nil {
			b.logger.Errorf("Failed to get student ID: %v", err)
			return nil, err
		}

		groupID, err := b.userRepo.GetStudentGroupID(studentID)
		if err != nil {
			b.logger.Errorf("Failed to get group ID for student %d: %v", studentID, err)
			return nil, err
		}

		return b.scheduleRepo.GetScheduleForDateByGroup(weekday, groupID)

	default:
		b.logger.Warnf("User %d with role %s tried to access schedule", maxUserID, userRole)
		return nil, fmt.Errorf("schedule not available for role: %s", userRole)
	}
}

func (b *Bot) sendScheduleForDay(ctx context.Context, maxUserID int64, callbackID string, weekday int16) error {
	entries, err := b.getScheduleEntriesForUser(maxUserID, weekday)
	if err != nil {
		return err
	}

	text := b.formatSchedule(entries, weekday)
	b.logger.Infof("Sending schedule for weekday %d to user %d", weekday, maxUserID)

	prevDay, nextDay := b.calculateNavigationDays(weekday)

	messageBody := &schemes.NewMessageBody{
		Text:        text,
		Format:      "markdown",
		Attachments: []interface{}{schemes.NewInlineKeyboardAttachmentRequest(GetScheduleKeyboard(b.MaxAPI, prevDay, nextDay).Build())},
	}

	answer := &schemes.CallbackAnswer{Message: messageBody}
	_, err = b.MaxAPI.Messages.AnswerOnCallback(ctx, callbackID, answer)
	if err != nil && err.Error() != "" {
		b.logger.Errorf("Failed to answer callback with schedule: %v", err)
		return err
	}

	return nil
}

func (b *Bot) answerScheduleCallback(ctx context.Context, maxUserID int64, callbackID string, weekday int16) error {
	entries, err := b.getScheduleEntriesForUser(maxUserID, weekday)
	if err != nil {
		return err
	}

	text := b.formatSchedule(entries, weekday)
	b.logger.Infof("Answering callback for weekday %d, user %d", weekday, maxUserID)

	prevDay, nextDay := b.calculateNavigationDays(weekday)

	if err := b.answerCallbackWithKeyboard(ctx, callbackID, GetScheduleKeyboard(b.MaxAPI, prevDay, nextDay), text); err != nil {
		b.logger.Errorf("Failed to answer callback: %v", err)
		return err
	}

	return nil
}
