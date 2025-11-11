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

	studentsPerPage = 5
)

func (b *Bot) handleMarkGradeStart(ctx context.Context, userID int64, callbackID string) error {
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
		b.answerCallbackWithNotification(ctx, callbackID, "У вас нет предметов для выставления оценок.")
		return nil
	}

	keyboard := b.MaxAPI.Messages.NewKeyboardBuilder()
	for _, subject := range subjects {
		payload := fmt.Sprintf("grade_subj_%d", subject.SubjectID)
		keyboard.AddRow().AddCallback(subject.SubjectName, schemes.DEFAULT, payload)
	}

	keyboard.AddRow().AddCallback(btnBackToMenu, schemes.DEFAULT, payloadBackToMenu)

	messageBody := &schemes.NewMessageBody{
		Text:        selectSubjectMsg,
		Attachments: []interface{}{schemes.NewInlineKeyboardAttachmentRequest(keyboard.Build())},
	}

	answer := &schemes.CallbackAnswer{Message: messageBody}
	_, err = b.MaxAPI.Messages.AnswerOnCallback(ctx, callbackID, answer)
	return err
}

func (b *Bot) handleGradeCallback(ctx context.Context, userID int64, callbackID, payload string) error {
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
		if len(parts) >= 3 && parts[2] == "page" {
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

func (b *Bot) handleStudentPageNavigation(ctx context.Context, _ int64, callbackID, payload string) error {
	var subjectID, groupID int64
	var page, totalPages int
	fmt.Sscanf(payload, "grade_stud_page_%d_%d_%d_%d", &subjectID, &groupID, &page, &totalPages)

	students, err := b.gradeRepo.GetStudentsByGroup(groupID)
	if err != nil {
		b.logger.Errorf("Failed to get students: %v", err)
		return err
	}

	return b.showStudentsPage(ctx, callbackID, subjectID, groupID, page, students)
}

func (b *Bot) handleSubjectSelected(ctx context.Context, userID int64, callbackID, payload string) error {
	var subjectID int64
	fmt.Sscanf(payload, "grade_subj_%d", &subjectID)

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
		b.answerCallbackWithNotification(ctx, callbackID, "У данного предмета нет групп.")
		return nil
	}

	keyboard := b.MaxAPI.Messages.NewKeyboardBuilder()
	for _, group := range groups {
		payload := fmt.Sprintf("grade_grp_%d_%d", subjectID, group.GroupID)
		keyboard.AddRow().AddCallback(group.GroupName, schemes.DEFAULT, payload)
	}

	keyboard.AddRow().AddCallback(btnBackToMenu, schemes.DEFAULT, payloadBackToMenu)

	messageBody := &schemes.NewMessageBody{
		Text:        selectGroupMsg,
		Attachments: []interface{}{schemes.NewInlineKeyboardAttachmentRequest(keyboard.Build())},
	}

	answer := &schemes.CallbackAnswer{Message: messageBody}
	_, err = b.MaxAPI.Messages.AnswerOnCallback(ctx, callbackID, answer)
	return err
}

func (b *Bot) handleGroupSelected(ctx context.Context, _ int64, callbackID, payload string) error {
	var subjectID, groupID int64
	fmt.Sscanf(payload, "grade_grp_%d_%d", &subjectID, &groupID)

	students, err := b.gradeRepo.GetStudentsByGroup(groupID)
	if err != nil {
		b.logger.Errorf("Failed to get students: %v", err)
		return err
	}

	if len(students) == 0 {
		b.answerCallbackWithNotification(ctx, callbackID, "В группе нет студентов.")
		return nil
	}

	return b.showStudentsPage(ctx, callbackID, subjectID, groupID, 0, students)
}

func (b *Bot) showStudentsPage(ctx context.Context, callbackID string, subjectID, groupID int64, page int, students []database.User) error {
	totalPages := (len(students) + studentsPerPage - 1) / studentsPerPage

	startIdx := page * studentsPerPage
	endIdx := startIdx + studentsPerPage
	if endIdx > len(students) {
		endIdx = len(students)
	}

	pageStudents := students[startIdx:endIdx]

	keyboard := GetStudentsPaginationKeyboard(b.MaxAPI, subjectID, groupID, page, totalPages, pageStudents)

	text := fmt.Sprintf(selectStudentMsg, page+1, totalPages)

	messageBody := &schemes.NewMessageBody{
		Text:        text,
		Attachments: []interface{}{schemes.NewInlineKeyboardAttachmentRequest(keyboard.Build())},
	}

	answer := &schemes.CallbackAnswer{Message: messageBody}
	_, err := b.MaxAPI.Messages.AnswerOnCallback(ctx, callbackID, answer)
	return err
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
		b.answerCallbackWithNotification(ctx, callbackID, "Нет расписания для данной группы.")
		return nil
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
	messageBody := &schemes.NewMessageBody{
		Text:        selectScheduleMsg,
		Attachments: []interface{}{schemes.NewInlineKeyboardAttachmentRequest(keyboard.Build())},
	}

	answer := &schemes.CallbackAnswer{Message: messageBody}
	_, err = b.MaxAPI.Messages.AnswerOnCallback(ctx, callbackID, answer)
	return err
}

func (b *Bot) handleScheduleSelected(ctx context.Context, _ int64, callbackID, payload string) error {
	var scheduleID, studentID int64
	fmt.Sscanf(payload, "grade_sch_%d_%d", &scheduleID, &studentID)

	var studentName string
	err := b.db.Get(&studentName, `SELECT name FROM users WHERE user_id = $1`, studentID)
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
	messageBody := &schemes.NewMessageBody{
		Text:        fmt.Sprintf(selectGradeMsg, studentName),
		Format:      "markdown",
		Attachments: []interface{}{schemes.NewInlineKeyboardAttachmentRequest(keyboard.Build())},
	}

	answer := &schemes.CallbackAnswer{Message: messageBody}
	_, err = b.MaxAPI.Messages.AnswerOnCallback(ctx, callbackID, answer)
	return err
}

func (b *Bot) handleGradeValueSelected(ctx context.Context, userID int64, callbackID, payload string) error {
	var scheduleID, studentID int64
	var gradeValue int
	fmt.Sscanf(payload, "grade_val_%d_%d_%d", &scheduleID, &studentID, &gradeValue)

	teacherID, err := b.userRepo.GetUserIDByMaxID(userID)
	if err != nil {
		return err
	}

	var subjectID int64
	err = b.db.Get(&subjectID, `SELECT subject_id FROM schedule WHERE schedule_id = $1`, scheduleID)
	if err != nil {
		b.logger.Errorf("Failed to get subject_id: %v", err)
		return err
	}

	err = b.gradeRepo.CreateGrade(studentID, teacherID, subjectID, scheduleID, gradeValue)
	if err != nil {
		b.logger.Errorf("Failed to create grade: %v", err)
		b.answerCallbackWithNotification(ctx, callbackID, "Ошибка при сохранении оценки.")
		return err
	}

	var studentName, subjectName string
	b.db.Get(&studentName, `SELECT name FROM users WHERE user_id = $1`, studentID)
	b.db.Get(&subjectName, `SELECT subject_name FROM subjects WHERE subject_id = $1`, subjectID)

	successText := fmt.Sprintf(gradeSuccessMsg, gradeValue, studentName, subjectName)

	userRole, err := b.getUserRole(userID)
	if err != nil {
		b.logger.Errorf("Failed to get user role: %v", err)
		userRole = "teacher"
	}

	var keyboard *maxbot.Keyboard
	switch userRole {
	case "teacher":
		keyboard = GetTeacherKeyboard(b.MaxAPI)
	default:
		keyboard = GetTeacherKeyboard(b.MaxAPI)
	}

	messageBody := &schemes.NewMessageBody{
		Text:        successText,
		Format:      "markdown",
		Attachments: []interface{}{schemes.NewInlineKeyboardAttachmentRequest(keyboard.Build())},
	}

	answer := &schemes.CallbackAnswer{
		Message:      messageBody,
		Notification: "Оценка выставлена!",
	}

	_, err = b.MaxAPI.Messages.AnswerOnCallback(ctx, callbackID, answer)
	return err
}

func (b *Bot) answerCallbackWithNotification(ctx context.Context, callbackID, notification string) error {
	answer := &schemes.CallbackAnswer{
		Notification: notification,
	}
	_, err := b.MaxAPI.Messages.AnswerOnCallback(ctx, callbackID, answer)
	return err
}
