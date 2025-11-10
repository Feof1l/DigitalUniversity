package maxAPI

import (
	"context"
	"fmt"
	"strings"

	"digitalUniversity/database"

	"github.com/max-messenger/max-bot-api-client-go/schemes"
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

func (b *Bot) sendScheduleForDay(ctx context.Context, u *schemes.MessageCallbackUpdate, maxUserID int64, weekday int16) error {
	userRole, err := b.getUserRole(maxUserID)
	if err != nil {
		b.logger.Errorf("Failed to get user role: %v", err)
		return err
	}

	var entries []database.Schedule

	if userRole == "teacher" {
		teacherID, err := b.userRepo.GetUserIDByMaxID(maxUserID)
		if err != nil {
			b.logger.Errorf("Failed to get teacher ID: %v", err)
			return err
		}
		entries, err = b.scheduleRepo.GetScheduleForDateByTeacher(weekday, teacherID)
		if err != nil {
			b.logger.Errorf("Failed to get schedule for teacher %d, weekday %d: %v", teacherID, weekday, err)
			return err
		}
	} else {
		entries, err = b.scheduleRepo.GetScheduleForDate(weekday)
		if err != nil {
			b.logger.Errorf("Failed to get schedule for weekday %d: %v", weekday, err)
			return err
		}
	}

	text := b.formatSchedule(entries, weekday)
	b.logger.Infof("Sending schedule for weekday %d to user %d", weekday, maxUserID)

	prevDay, nextDay := b.calculateNavigationDays(weekday)

	//Попытки редактировать сообщение, может быть будет полезно
	// b.logger.Infof("aaaaaaaaaaaaaaaaa %s: %v %v", b.scheduleMessageIDs[maxUserID], maxUserID, u.Message.Body.Mid)
	// if msgID, exists := b.scheduleMessageIDs[maxUserID]; exists {

	// 	mID, err := strconv.Atoi(msgID)
	// 	if err != nil {
	// 		b.logger.Errorf("failed to convert msgID %v", err)
	// 	}
	// 	err = b.MaxAPI.Messages.EditMessage(ctx, int64(mID), maxbot.NewMessage().
	// 		SetUser(maxUserID).
	// 		AddKeyboard(GetScheduleKeyboard(b.MaxAPI, prevDay, nextDay)).
	// 		SetText(text))

	// 	if err != nil {
	// 		b.logger.Errorf("Failed to edit message %s for user %d: %v", msgID, maxUserID, err)
	// 		delete(b.scheduleMessageIDs, maxUserID)
	// 		b.sendKeyboard(ctx, GetScheduleKeyboard(b.MaxAPI, prevDay, nextDay), maxUserID, text)

	// 		//return b.sendNewScheduleMessage(ctx, maxUserID, weekday, text, keyboard)
	// 	}
	// 	return nil
	// } else {
	// 	//delete(b.scheduleMessageIDs, maxUserID)

	// 	//b.sendKeyboard(ctx, GetScheduleKeyboard(b.MaxAPI, prevDay, nextDay), maxUserID, text)

	// 	id, err := b.MaxAPI.Messages.Send(ctx, maxbot.NewMessage().
	// 		SetUser(maxUserID).
	// 		AddKeyboard(GetScheduleKeyboard(b.MaxAPI, prevDay, nextDay)).
	// 		SetText(text))
	// 	b.logger.Errorf("dsdsds %s: %v %v %v", msgID, maxUserID, u.Message.Body.Mid, id)

	// 	if err != nil && err.Error() != "" {
	// 		b.logger.Errorf("Failed to send keyboard: %v", err)
	// 	}
	// 	ids:=u.Message.Body.

	// 	b.scheduleMessageIDs[maxUserID] = u.Message.Body.Mid
	// }

	b.sendKeyboard(ctx, GetScheduleKeyboard(b.MaxAPI, prevDay, nextDay), maxUserID, text)

	return nil
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
