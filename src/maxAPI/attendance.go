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
	selectSubjectForAttendanceMsg  = "Выберите предмет для отметки посещаемости:"
	selectGroupForAttendanceMsg    = "Выберите группу:"
	selectScheduleForAttendanceMsg = "Выберите занятие:"
	selectAbsentStudentsMsg        = "Отметьте посещаемость на занятии:\n**%s %s**\n\nВыберите отсутствующих:"
	allMarkedPresentMsg            = "✅ Студенты отмечены как присутствующие!"

	attendanceStatsHeaderMsg = "📊 Посещаемость по предмету **%s**:\n\n"
	attendanceEntryFormat    = "`%s %s` — %s\n"
	attendanceStatsFooter    = "\n📈 **Статистика:**\n" +
		"• Всего занятий: **%d**\n" +
		"• Посещено: **%d** (%d%%)\n" +
		"• Пропущено: **%d**"

	btnMarkAllPresent = "✅ Отметить всех остальных присутствующими"
)

func (b *Bot) handleMarkAttendanceStart(ctx context.Context, userID int64, callbackID string) error {
	teacherID, err := b.userRepo.GetUserIDByMaxID(userID)
	if err != nil {
		b.logger.Errorf("Failed to get teacher ID: %v", err)
		return err
	}

	subjects, err := b.gradeRepo.GetSubjectsByTeacher(teacherID)
	if err != nil {
		b.logger.Errorf("Failed to get subjects for teacher %d: %v", teacherID, err)
		return err
	}

	if len(subjects) == 0 {
		return b.answerCallbackWithNotification(ctx, callbackID, "У вас нет предметов для отметки посещаемости.")
	}

	keyboard := b.MaxAPI.Messages.NewKeyboardBuilder()
	for _, subject := range subjects {
		payload := fmt.Sprintf("attend_subj_%d", subject.SubjectID)
		keyboard.AddRow().AddCallback(subject.SubjectName, schemes.DEFAULT, payload)
	}
	keyboard.AddRow().AddCallback(btnBackToMenu, schemes.DEFAULT, payloadBackToMenu)

	return b.answerWithKeyboard(ctx, callbackID, selectSubjectForAttendanceMsg, keyboard)
}

func (b *Bot) handleAttendanceCallback(ctx context.Context, userID int64, callbackID, payload string) error {
	if !strings.HasPrefix(payload, "attend_") {
		return fmt.Errorf("invalid attendance callback payload: %s", payload)
	}

	parts := strings.Split(payload, "_")
	if len(parts) < 2 {
		return fmt.Errorf("invalid attendance callback payload: %s", payload)
	}

	switch parts[1] {
	case "subj":
		return b.handleAttendanceSubjectSelected(ctx, userID, callbackID, payload)
	case "grp":
		return b.handleAttendanceGroupSelected(ctx, userID, callbackID, payload)
	case "sch":
		return b.handleAttendanceScheduleSelected(ctx, userID, callbackID, payload)
	case "all":
		return b.handleAttendanceMarkAll(ctx, userID, callbackID, payload)
	case "absent":
		return b.handleAttendanceMarkAbsent(ctx, userID, callbackID, payload)
	default:
		return fmt.Errorf("unknown attendance callback type: %s", parts[1])
	}
}

func (b *Bot) handleAttendanceSubjectSelected(ctx context.Context, userID int64, callbackID, payload string) error {
	var subjectID int64
	fmt.Sscanf(payload, "attend_subj_%d", &subjectID)

	teacherID, err := b.userRepo.GetUserIDByMaxID(userID)
	if err != nil {
		return err
	}

	groups, err := b.gradeRepo.GetGroupsBySubjectAndTeacher(subjectID, teacherID)
	if err != nil {
		b.logger.Errorf("Failed to get groups: %v", err)
		return err
	}

	if len(groups) == 0 {
		return b.answerCallbackWithNotification(ctx, callbackID, "У данного предмета нет групп.")
	}

	keyboard := b.MaxAPI.Messages.NewKeyboardBuilder()
	for _, group := range groups {
		payload := fmt.Sprintf("attend_grp_%d_%d", subjectID, group.GroupID)
		keyboard.AddRow().AddCallback(group.GroupName, schemes.DEFAULT, payload)
	}
	keyboard.AddRow().AddCallback(btnBackToMenu, schemes.DEFAULT, payloadBackToMenu)

	return b.answerWithKeyboard(ctx, callbackID, selectGroupForAttendanceMsg, keyboard)
}

func (b *Bot) handleAttendanceGroupSelected(ctx context.Context, _ int64, callbackID, payload string) error {
	var subjectID, groupID int64
	fmt.Sscanf(payload, "attend_grp_%d_%d", &subjectID, &groupID)

	schedules, err := b.gradeRepo.GetScheduleBySubjectAndGroup(subjectID, groupID)
	if err != nil {
		b.logger.Errorf("Failed to get schedules: %v", err)
		return err
	}

	if len(schedules) == 0 {
		return b.answerCallbackWithNotification(ctx, callbackID, "Нет расписания для данной группы.")
	}

	keyboard := b.MaxAPI.Messages.NewKeyboardBuilder()
	for _, schedule := range schedules {
		dayName := b.getWeekdayName(schedule.Weekday)
		startTime := schedule.StartTime.Format("15:04")
		btnText := fmt.Sprintf("%s %s", dayName, startTime)
		payload := fmt.Sprintf("attend_sch_%d_%d_%d", subjectID, groupID, schedule.ScheduleID)
		keyboard.AddRow().AddCallback(btnText, schemes.DEFAULT, payload)
	}
	keyboard.AddRow().AddCallback(btnBackToMenu, schemes.DEFAULT, payloadBackToMenu)

	return b.answerWithKeyboard(ctx, callbackID, selectScheduleForAttendanceMsg, keyboard)
}

func (b *Bot) handleAttendanceScheduleSelected(ctx context.Context, _ int64, callbackID, payload string) error {
	var subjectID, groupID, scheduleID int64
	fmt.Sscanf(payload, "attend_sch_%d_%d_%d", &subjectID, &groupID, &scheduleID)

	students, err := b.gradeRepo.GetStudentsByGroup(groupID)
	if err != nil {
		b.logger.Errorf("Failed to get students: %v", err)
		return err
	}

	if len(students) == 0 {
		b.answerCallbackWithNotification(ctx, callbackID, "В группе нет студентов.")
		return nil
	}

	var weekday int16
	var startTime time.Time
	b.db.Get(&weekday, `SELECT weekday FROM schedule WHERE schedule_id = $1`, scheduleID)
	b.db.Get(&startTime, `SELECT start_time FROM schedule WHERE schedule_id = $1`, scheduleID)

	dayName := b.getWeekdayName(weekday)
	timeStr := startTime.Format("15:04")

	return b.showAttendanceStudentsList(ctx, callbackID, subjectID, groupID, scheduleID, dayName, timeStr, students, []int64{})
}

func (b *Bot) showAttendanceStudentsList(ctx context.Context, callbackID string, subjectID, groupID, scheduleID int64, dayName, timeStr string, allStudents []database.User, markedAbsentIDs []int64) error {
	var availableStudents []database.User
	for _, student := range allStudents {
		isMarked := false
		for _, markedID := range markedAbsentIDs {
			if student.UserID == markedID {
				isMarked = true
				break
			}
		}
		if !isMarked {
			availableStudents = append(availableStudents, student)
		}
	}

	if len(availableStudents) == 0 {
		keyboard := GetTeacherKeyboard(b.MaxAPI)
		return b.answerWithKeyboard(ctx, callbackID, allMarkedPresentMsg, keyboard)
	}

	text := fmt.Sprintf(selectAbsentStudentsMsg, dayName, timeStr)

	keyboard := b.MaxAPI.Messages.NewKeyboardBuilder()

	payload := fmt.Sprintf("attend_all_%d_%d_%d", subjectID, groupID, scheduleID)
	keyboard.AddRow().AddCallback(btnMarkAllPresent, schemes.POSITIVE, payload)

	for _, student := range availableStudents {
		markedIDsStr := ""
		for _, id := range markedAbsentIDs {
			if markedIDsStr != "" {
				markedIDsStr += ","
			}
			markedIDsStr += fmt.Sprintf("%d", id)
		}
		if markedIDsStr == "" {
			markedIDsStr = "0"
		}

		payload := fmt.Sprintf("attend_absent_%d_%d_%d_%d_%s", subjectID, groupID, scheduleID, student.UserID, markedIDsStr)
		keyboard.AddRow().AddCallback(student.Name, schemes.DEFAULT, payload)
	}

	keyboard.AddRow().AddCallback(btnBackToMenu, schemes.DEFAULT, payloadBackToMenu)

	return b.answerWithKeyboardMarkdown(ctx, callbackID, text, keyboard)
}

func (b *Bot) handleAttendanceMarkAll(ctx context.Context, _ int64, callbackID, payload string) error {
	var subjectID, groupID, scheduleID int64
	fmt.Sscanf(payload, "attend_all_%d_%d_%d", &subjectID, &groupID, &scheduleID)

	students, err := b.gradeRepo.GetStudentsByGroup(groupID)
	if err != nil {
		b.logger.Errorf("Failed to get students: %v", err)
		return err
	}

	attendanceRecords, err := b.attendanceRepo.GetAttendanceRecordsBySchedule(scheduleID)
	if err != nil {
		b.logger.Errorf("Failed to get attendance records: %v", err)
		return err
	}

	attendanceMap := make(map[int64]bool)
	for _, record := range attendanceRecords {
		attendanceMap[record.StudentID] = record.Attended
	}

	subjectName, _ := b.subjectRepo.GetSubjectName(subjectID)
	now := time.Now()

	for _, student := range students {
		attended, exists := attendanceMap[student.UserID]
		if !exists {
			err := b.attendanceRepo.MarkAttendance(student.UserID, scheduleID, true)
			if err != nil {
				b.logger.Errorf("Failed to mark attendance for student %d: %v", student.UserID, err)
			} else {
				go b.notifyStudentAttendance(context.Background(), student.UserID, subjectID, subjectName, true, now)
			}
		} else if !attended {
			continue
		}
	}

	keyboard := GetTeacherKeyboard(b.MaxAPI)
	return b.answerWithKeyboardAndNotification(ctx, callbackID, allMarkedPresentMsg, keyboard, "Все отмечены!")
}

func (b *Bot) handleAttendanceMarkAbsent(ctx context.Context, _ int64, callbackID, payload string) error {
	var subjectID, groupID, scheduleID, studentID int64
	var markedIDsStr string

	fmt.Sscanf(payload, "attend_absent_%d_%d_%d_%d_%s", &subjectID, &groupID, &scheduleID, &studentID, &markedIDsStr)

	var markedAbsentIDs []int64
	if markedIDsStr != "0" {
		parts := strings.Split(markedIDsStr, ",")
		for _, part := range parts {
			var id int64
			fmt.Sscanf(part, "%d", &id)
			if id != 0 {
				markedAbsentIDs = append(markedAbsentIDs, id)
			}
		}
	}

	err := b.attendanceRepo.MarkAttendance(studentID, scheduleID, false)
	if err != nil {
		b.logger.Errorf("Failed to mark attendance: %v", err)
		b.answerCallbackWithNotification(ctx, callbackID, "Ошибка при отметке посещаемости.")
		return err
	}
	subjectName, _ := b.subjectRepo.GetSubjectName(subjectID)
	go b.notifyStudentAttendance(context.Background(), studentID, subjectID, subjectName, false, time.Now())

	markedAbsentIDs = append(markedAbsentIDs, studentID)

	students, err := b.gradeRepo.GetStudentsByGroup(groupID)
	if err != nil {
		b.logger.Errorf("Failed to get students: %v", err)
		return err
	}

	var weekday int16
	var startTime time.Time
	b.db.Get(&weekday, `SELECT weekday FROM schedule WHERE schedule_id = $1`, scheduleID)
	b.db.Get(&startTime, `SELECT start_time FROM schedule WHERE schedule_id = $1`, scheduleID)

	dayName := b.getWeekdayName(weekday)
	timeStr := startTime.Format("15:04")

	return b.showAttendanceStudentsList(ctx, callbackID, subjectID, groupID, scheduleID, dayName, timeStr, students, markedAbsentIDs)
}

func (b *Bot) handleShowAttendanceStart(ctx context.Context, userID int64, callbackID string) error {
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
		return b.answerCallbackWithNotification(ctx, callbackID, "У вашей группы пока нет предметов.")
	}

	keyboard := b.MaxAPI.Messages.NewKeyboardBuilder()
	for _, subject := range subjects {
		buttonText := subject.SubjectName
		payload := fmt.Sprintf("show_attend_subj_%d", subject.SubjectID)
		keyboard.AddRow().AddCallback(buttonText, schemes.DEFAULT, payload)
	}
	keyboard.AddRow().AddCallback(btnBackToMenu, schemes.DEFAULT, payloadBackToMenu)

	return b.answerWithKeyboard(ctx, callbackID, "Выберите предмет для просмотра посещаемости:", keyboard)
}

func (b *Bot) handleShowAttendanceCallback(ctx context.Context, userID int64, callbackID, payload string) error {
	if !strings.HasPrefix(payload, "show_attend_") {
		return fmt.Errorf("invalid show attendance callback payload: %s", payload)
	}

	parts := strings.Split(payload, "_")
	if len(parts) < 4 {
		return fmt.Errorf("invalid show attendance callback payload format: %s", payload)
	}

	if parts[2] == "subj" {
		return b.handleShowAttendanceSubjectSelected(ctx, userID, callbackID, payload)
	}

	return fmt.Errorf("unknown show attendance callback type: %s", payload)
}

func (b *Bot) handleShowAttendanceSubjectSelected(ctx context.Context, userID int64, callbackID, payload string) error {
	var subjectID int64
	fmt.Sscanf(payload, "show_attend_subj_%d", &subjectID)

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

	attendance, err := b.attendanceRepo.GetAttendanceByStudentAndSubject(studentID, subjectID)
	if err != nil {
		b.logger.Errorf("Failed to get attendance: %v", err)
		return err
	}

	text := b.formatAttendanceList(attendance, subjectName)

	keyboard := GetStudentKeyboard(b.MaxAPI)

	return b.answerWithKeyboardMarkdown(ctx, callbackID, text, keyboard)
}

func (b *Bot) formatAttendanceList(attendance []database.Attendance, subjectName string) string {
	if len(attendance) == 0 {
		return fmt.Sprintf("По предмету **%s** посещаемость пока не отмечена.", subjectName)
	}

	var sb strings.Builder
	fmt.Fprintf(&sb, attendanceStatsHeaderMsg, subjectName)

	var present, total int

	for _, att := range attendance {
		moscowTime := att.MarkTime.Add(3 * time.Hour)
		dateStr := moscowTime.Format("02.01.2006")
		timeStr := moscowTime.Format("15:04:05")

		status := "❌ Не был"
		if att.Attended {
			status = "✅ Был"
			present++
		}
		total++

		fmt.Fprintf(&sb, attendanceEntryFormat, dateStr, timeStr, status)
	}

	percentage := 0
	if total > 0 {
		percentage = (present * 100) / total
	}
	absent := total - present

	fmt.Fprintf(&sb, attendanceStatsFooter, total, present, percentage, absent)

	return sb.String()
}

func (b *Bot) notifyStudentAttendance(ctx context.Context, studentID, subjectID int64, subjectName string, attended bool, markTime time.Time) {
	dateStr := markTime.Add(3 * time.Hour).Format("02.01.2006 15:04")
	statusEmoji := "✅ Был"
	if !attended {
		statusEmoji = "❌ Не был"
	}

	notificationText := fmt.Sprintf(
		"📝 **Посещаемость**\n\n"+
			"Предмет: **%s**\n"+
			"Дата: `%s`\n"+
			"Статус: %s",
		subjectName, dateStr, statusEmoji,
	)

	keyboard := b.MaxAPI.Messages.NewKeyboardBuilder()
	payload := fmt.Sprintf("show_attend_subj_%d", subjectID)
	keyboard.AddRow().AddCallback("📊 Посмотреть статистику посещаемости", schemes.POSITIVE, payload)

	studentMaxID, err := b.userRepo.GetUserMaxIDByID(studentID)
	if err != nil {
		b.logger.Warnf("Failed to get student max_id for attendance notification: %v", err)
		return
	}

	b.sendKeyboard(ctx, keyboard, studentMaxID, notificationText)
}
