package maxAPI

import (
	"context"
	"fmt"
	"strings"

	maxbot "github.com/max-messenger/max-bot-api-client-go"
	"github.com/max-messenger/max-bot-api-client-go/schemes"

	"digitalUniversity/database"
)

const (
	selectSubjectMsg  = "Выберите предмет:"
	selectGroupMsg    = "Выберите группу:"
	selectScheduleMsg = "Выберите занятие:"
	selectStudentMsg  = "Выберите студента (страница %d/%d):"
	selectGradeMsg    = "Выберите оценку для студента **%s**:"
	gradeSuccessMsg   = "✅ Оценка **%d** успешно выставлена студенту **%s** по предмету **%s**!"

	noGroupsMsg       = "У данного предмета нет групп."
	noStudentsMsg     = "В группе нет студентов."
	noSubjectsMsg     = "У вас нет предметов для выставления оценок."
	noScheduleMsg     = "Нет расписания для данной группы."
	gradeSaveErrorMsg = "Ошибка при сохранении оценки."

	notificationTextTemplate = "📚 **Новая оценка!**\n\nПредмет: **%s**\nОценка: **%d**"

	studentsPerPage = 5
)

func (b *Bot) handleMarkGradeStart(ctx context.Context, userID int64, callbackID string) error {
	teacherID, err := b.userRepo.GetUserIDByMaxID(userID)
	if err != nil {
		b.logger.Errorf("Failed to get teacher ID: %v", err)
		return err
	}

	var subjects []database.Subject

	if b.superUser[userID] {
		subjects, err = b.gradeRepo.GetSubjects()
		if err != nil {
			b.logger.Errorf("Failed to get subjects %v", err)
			return err
		}
	} else {
		subjects, err = b.gradeRepo.GetSubjectsByTeacher(teacherID)
		if err != nil {
			b.logger.Errorf("Failed to get subjects for teacher %d: %v", teacherID, err)
			return err
		}
	}

	if len(subjects) == 0 {
		return b.answerCallbackWithNotification(ctx, callbackID, noSubjectsMsg)
	}

	keyboard := b.MaxAPI.Messages.NewKeyboardBuilder()
	for _, subject := range subjects {
		payload := fmt.Sprintf("grade_subj_%d", subject.SubjectID)
		keyboard.AddRow().AddCallback(subject.SubjectName, schemes.DEFAULT, payload)
	}
	keyboard.AddRow().AddCallback(btnBackToMenu, schemes.DEFAULT, payloadBackToMenu)

	return b.answerWithKeyboard(ctx, callbackID, selectSubjectMsg, keyboard)
}

func (b *Bot) handleGradeCallback(ctx context.Context, userID int64, callbackID, payload string) error {
	if !strings.HasPrefix(payload, "grade_") {
		return fmt.Errorf("invalid grade callback payload: %s", payload)
	}

	parts := strings.Split(payload, "_")
	if len(parts) < 2 {
		return fmt.Errorf("invalid grade callback payload: %s", payload)
	}

	switch parts[1] {
	case "subj":
		return b.handleSubjectSelected(ctx, userID, callbackID, payload)
	case "grp":
		return b.handleGroupSelected(ctx, userID, callbackID, payload)
	case "stud":
		if len(parts) >= 4 && parts[2] == "page" {
			return b.handleStudentPageNavigation(ctx, userID, callbackID, payload)
		}
		return b.handleStudentSelected(ctx, userID, callbackID, payload)
	case "sch":
		return b.handleScheduleSelected(ctx, userID, callbackID, payload)
	case "val":
		return b.handleGradeValueSelected(ctx, userID, callbackID, payload)
	default:
		return fmt.Errorf("unknown grade callback type: %s", parts[1])
	}
}

func (b *Bot) handleStudentPageNavigation(ctx context.Context, userID int64, callbackID, payload string) error {
	var subjectID, groupID int64
	var page, totalPages int
	n, err := fmt.Sscanf(payload, "grade_stud_page_%d_%d_%d_%d", &subjectID, &groupID, &page, &totalPages)
	if n != 4 || err != nil {
		b.logger.Errorf("invalid page navigation payload: %s", payload)
		return err
	}

	students, err := b.gradeRepo.GetStudentsByGroup(groupID)
	if err != nil {
		b.logger.Errorf("Failed to get students: %v", err)
		return err
	}

	return b.showStudentsPage(ctx, userID, callbackID, subjectID, groupID, page, students)
}

func (b *Bot) handleSubjectSelected(ctx context.Context, userID int64, callbackID, payload string) error {
	var subjectID int64
	fmt.Sscanf(payload, "grade_subj_%d", &subjectID)

	teacherID, err := b.userRepo.GetUserIDByMaxID(userID)
	if err != nil {
		return err
	}

	var groups []database.Group

	if b.superUser[userID] {
		groups, err = b.gradeRepo.GetGroupsBySubject(subjectID)
	} else {
		groups, err = b.gradeRepo.GetGroupsBySubjectAndTeacher(subjectID, teacherID)
	}
	if err != nil {
		b.logger.Errorf("Failed to get groups: %v", err)
		return err
	}

	if len(groups) == 0 {
		return b.answerCallbackWithNotification(ctx, callbackID, noGroupsMsg)
	}

	keyboard := b.MaxAPI.Messages.NewKeyboardBuilder()
	for _, group := range groups {
		payload := fmt.Sprintf("grade_grp_%d_%d", subjectID, group.GroupID)
		keyboard.AddRow().AddCallback(group.GroupName, schemes.DEFAULT, payload)
	}
	keyboard.AddRow().AddCallback(btnBackToMenu, schemes.DEFAULT, payloadBackToMenu)

	return b.answerWithKeyboard(ctx, callbackID, selectGroupMsg, keyboard)
}

func (b *Bot) handleGroupSelected(ctx context.Context, userID int64, callbackID, payload string) error {
	var subjectID, groupID int64
	fmt.Sscanf(payload, "grade_grp_%d_%d", &subjectID, &groupID)

	students, err := b.gradeRepo.GetStudentsByGroup(groupID)
	if err != nil {
		b.logger.Errorf("Failed to get students: %v", err)
		return err
	}

	if len(students) == 0 {
		return b.answerCallbackWithNotification(ctx, callbackID, noStudentsMsg)
	}

	return b.showStudentsPage(ctx, userID, callbackID, subjectID, groupID, 0, students)
}

func (b *Bot) showStudentsPage(ctx context.Context, userID int64, callbackID string, subjectID, groupID int64, page int, students []database.User) error {
	totalPages := (len(students) + studentsPerPage - 1) / studentsPerPage

	startIdx := page * studentsPerPage
	endIdx := startIdx + studentsPerPage
	if endIdx > len(students) {
		endIdx = len(students)
	}

	pageStudents := students[startIdx:endIdx]

	keyboard := GetStudentsPaginationKeyboard(b.MaxAPI, subjectID, groupID, page, totalPages, pageStudents, b.superUser[userID])

	text := fmt.Sprintf(selectStudentMsg, page+1, totalPages)

	return b.answerWithKeyboard(ctx, callbackID, text, keyboard)
}

func (b *Bot) handleStudentSelected(ctx context.Context, _ int64, callbackID, payload string) error {
	var subjectID, groupID, studentID int64
	fmt.Sscanf(payload, "grade_stud_%d_%d_%d", &subjectID, &groupID, &studentID)

	schedules, err := b.gradeRepo.GetScheduleBySubjectAndGroup(subjectID, groupID)
	if err != nil {
		b.logger.Errorf("Failed to get schedules: %v", err)
		return err
	}

	if len(schedules) == 0 {
		return b.answerCallbackWithNotification(ctx, callbackID, noScheduleMsg)
	}

	keyboard := b.MaxAPI.Messages.NewKeyboardBuilder()
	for _, schedule := range schedules {
		dayName := b.getWeekdayName(schedule.Weekday)
		startTime := schedule.StartTime.Format("15:04")
		btnText := fmt.Sprintf("%s %s", dayName, startTime)
		payload := fmt.Sprintf("grade_sch_%d_%d", schedule.ScheduleID, studentID)
		keyboard.AddRow().AddCallback(btnText, schemes.DEFAULT, payload)
	}
	keyboard.AddRow().AddCallback(btnBackToMenu, schemes.DEFAULT, payloadBackToMenu)

	return b.answerWithKeyboard(ctx, callbackID, selectScheduleMsg, keyboard)
}

func (b *Bot) handleScheduleSelected(ctx context.Context, _ int64, callbackID, payload string) error {
	var scheduleID, studentID int64
	fmt.Sscanf(payload, "grade_sch_%d_%d", &scheduleID, &studentID)

	studentName, err := b.gradeRepo.GetStudentNameByID(studentID)
	if err != nil {
		b.logger.Errorf("Failed to get student name: %v", err)
		studentName = "студенту"
	}

	keyboard := b.MaxAPI.Messages.NewKeyboardBuilder()
	row := keyboard.AddRow()
	for grade := 0; grade <= 5; grade++ {
		payload := fmt.Sprintf("grade_val_%d_%d_%d", scheduleID, studentID, grade)
		row.AddCallback(fmt.Sprintf("%d", grade), schemes.DEFAULT, payload)
	}
	keyboard.AddRow().AddCallback(btnBackToMenu, schemes.DEFAULT, payloadBackToMenu)

	return b.answerWithKeyboard(ctx, callbackID, fmt.Sprintf(selectGradeMsg, studentName), keyboard)
}

func (b *Bot) handleGradeValueSelected(ctx context.Context, userID int64, callbackID, payload string) error {
	var scheduleID, studentID int64
	var gradeValue int
	fmt.Sscanf(payload, "grade_val_%d_%d_%d", &scheduleID, &studentID, &gradeValue)

	teacherID, err := b.userRepo.GetUserIDByMaxID(userID)
	if err != nil {
		return err
	}

	subjectID, err := b.gradeRepo.GetSubjectIDByScheduleID(scheduleID)
	if err != nil {
		b.logger.Errorf("Failed to get subject_id: %v", err)
		return err
	}

	err = b.gradeRepo.CreateGrade(studentID, teacherID, subjectID, scheduleID, gradeValue)
	if err != nil {
		b.logger.Errorf("Failed to create grade: %v", err)
		return b.answerCallbackWithNotification(ctx, callbackID, gradeSaveErrorMsg)
	}

	studentName, err := b.gradeRepo.GetStudentNameByID(studentID)
	if err != nil {
		b.logger.Warnf("Failed to get student name: %v", err)
		studentName = "студенту"
	}

	subjectName, err := b.subjectRepo.GetSubjectName(subjectID)
	if err != nil {
		b.logger.Warnf("Failed to get subject name: %v", err)
		subjectName = "предмету"
	}

	successText := fmt.Sprintf(gradeSuccessMsg, gradeValue, studentName, subjectName)

	keyboard := GetTeacherKeyboard(b.MaxAPI, b.superUser[userID])

	go b.sendGradeNotification(context.Background(), studentID, subjectID, subjectName, gradeValue)

	return b.answerWithKeyboardAndNotification(ctx, callbackID, successText, keyboard, "Оценка выставлена!")
}

func (b *Bot) sendGradeNotification(ctx context.Context, studentID, subjectID int64, subjectName string, gradeValue int) {
	studentMaxID, err := b.userRepo.GetUserMaxIDByID(studentID)
	if err != nil {
		b.logger.Warnf("Failed to get student max_id for notification: %v", err)
		return
	}

	notificationText := fmt.Sprintf(notificationTextTemplate, subjectName, gradeValue)

	keyboard := b.MaxAPI.Messages.NewKeyboardBuilder()
	payload := fmt.Sprintf("show_grades_subj_%d", subjectID)
	keyboard.AddRow().AddCallback("📊 Посмотреть все оценки", schemes.POSITIVE, payload)

	msg := maxbot.NewMessage().
		SetUser(studentMaxID).
		SetText(notificationText).
		SetFormat("markdown").
		AddKeyboard(keyboard)

	_, err = b.MaxAPI.Messages.Send(ctx, msg)
	if err != nil {
		b.logger.Warnf("Failed to send grade notification to student %d: %v", studentMaxID, err)
	} else {
		b.logger.Infof("Grade notification sent to student %d (max_id: %d)", studentID, studentMaxID)
	}
}

func (b *Bot) answerCallbackWithNotification(ctx context.Context, callbackID, notification string) error {
	answer := &schemes.CallbackAnswer{
		Notification: notification,
	}
	_, err := b.MaxAPI.Messages.AnswerOnCallback(ctx, callbackID, answer)
	return err
}
