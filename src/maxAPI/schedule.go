package maxAPI

import (
	"context"
	"fmt"
	"strings"

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

	if len(entries) == 0 {
		return fmt.Sprintf("📅 **%s**\n\nНет занятий.", dayName)
	}

	var sb strings.Builder
	fmt.Fprintf(&sb, "📅 **%s**\n\n", dayName)

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

func (b *Bot) getSubjectName(subjectID int64) string {
	name, err := b.subjectRepo.GetSubjectName(subjectID)
	if err != nil {
		b.logger.Errorf("Failed to get subject name for ID %d: %v", subjectID, err)
		return "Неизвестный предмет"
	}
	return name
}

func (b *Bot) getLessonTypeName(lessonTypeID int64) string {
	name, err := b.lessonTypeRepo.GetLessonTypeName(lessonTypeID)
	if err != nil {
		b.logger.Errorf("Failed to get lesson type name for ID %d: %v", lessonTypeID, err)
		return "Неизвестный тип"
	}
	return name
}

func (b *Bot) getTeacherName(teacherID int64) string {
	name, err := b.userRepo.GetTeacherName(teacherID)
	if err != nil {
		b.logger.Errorf("Failed to get teacher name for ID %d: %v", teacherID, err)
		return "Неизвестный преподаватель"
	}
	return name
}

func (b *Bot) getGroupName(groupID int64) string {
	name, err := b.groupRepo.GetGroupName(groupID)
	if err != nil {
		b.logger.Errorf("Failed to get group name for ID %d: %v", groupID, err)
		return "Неизвестная группа"
	}
	return name
}

func (b *Bot) sendScheduleForDay(ctx context.Context, maxUserID int64, weekday int16) error {
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
