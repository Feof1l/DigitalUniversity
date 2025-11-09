package maxAPI

import (
	"context"
	"digitalUniversity/database"
	"fmt"
	"strings"
)

func (b *Bot) formatSchedule(entries []database.Schedule, weekday int16) string {
	days := map[int16]string{
		1: "Понедельник",
		2: "Вторник",
		3: "Среда",
		4: "Четверг",
		5: "Пятница",
		6: "Суббота",
		7: "Воскресенье",
	}

	dayName := days[weekday]
	if dayName == "" {
		dayName = fmt.Sprintf("День %d", weekday)
	}

	if len(entries) == 0 {
		return fmt.Sprintf("📅 **%s**\n\nНет занятий.", dayName)
	}

	var s strings.Builder
	s.WriteString(fmt.Sprintf("📅 **%s**\n\n", dayName))

	userRepo := database.NewUserRepository(b.db)

	for i, e := range entries {
		subjectName, err := userRepo.GetSubjectName(e.SubjectID)
		if err != nil {
			b.logger.Errorf("Failed to get subjectName %v", err)
		}

		lessonTypeName, err := userRepo.GetLessonTypeName(e.LessonTypeID)
		if err != nil {
			b.logger.Errorf("Failed to get typeName %v", err)
		}

		teacherName, err := userRepo.GetTeacherName(e.TeacherID)
		if err != nil {
			b.logger.Errorf("Failed to get teacherName %v", err)
		}

		groupName, err := userRepo.GetGroupName(e.GroupID)
		if err != nil {
			b.logger.Errorf("Failed to get groupName %v", err)
		}

		start := e.StartTime.Format("15:04")
		end := e.EndTime.Format("15:04")
		s.WriteString(fmt.Sprintf(
			"%d. **%s** (%s)\n   👨‍🏫 %s\n  %s\n  🏫 %s\n   ⏰ %s–%s\n\n",
			i+1,
			subjectName,
			lessonTypeName,
			teacherName,
			groupName,
			e.ClassRoom,
			start,
			end,
		))
	}

	return strings.TrimSpace(s.String())
}

func (b *Bot) sendScheduleForDay(ctx context.Context, chatID int64, weekday int16) error {
	userRepo := database.NewUserRepository(b.db)
	entries, err := userRepo.GetScheduleForDate(weekday)
	if err != nil {
		b.logger.Errorf("sendScheduleForDay  %v", err)
		return err
	}

	text := b.formatSchedule(entries, weekday)

	b.logger.Infof("sendScheduleForDay  %v", text)

	prevDay := weekday - 1
	if prevDay < 1 {
		prevDay = 7
	}
	nextDay := weekday + 1
	if nextDay > 7 {
		nextDay = 1
	}

	b.sendKeyboard(ctx, GetScheduleKeyboard(b.MaxAPI, prevDay, nextDay), chatID, text)

	return err
}
