package maxAPI

import (
	"fmt"

	maxbot "github.com/max-messenger/max-bot-api-client-go"
	"github.com/max-messenger/max-bot-api-client-go/schemes"

	"digitalUniversity/database"
)

const (
	btnUploadStudents = "Загрузить файл со студентами"
	btnUploadTeachers = "Загрузить файл с преподавателями"
	btnUploadSchedule = "Загрузить файл с расписанием"

	btnShowSchedule   = "Показать расписание"
	btnMarkScore      = "Поставить оценку"
	btnMarkAttendance = "Отметить посещаемость"

	btnPrev           = "← Назад"
	btnNext           = "Вперёд →"
	btnBackToMenu     = "Главное меню"
	btnShowScore      = "Посмотреть оценки"
	btnShowAttendance = "Посмотреть посещаемость"

	btnAdmin   = "Администратор"
	btnTeacher = "Преподаватель"
	btnStudent = "Студент"

	btnBackToRoleSelection = "К выбору роли"

	payloadUploadStudents = "uploadStudents"
	payloadUploadTeachers = "uploadTeachers"
	payloadUploadSchedule = "uploadSchedule"
	payloadShowSchedule   = "showSchedule"
	payloadShowScore      = "showScore"
	payloadMarkGrade      = "markGrade"
	payloadMarkAttendance = "markAttendance"
	payloadShowAttendance = "showAttendance"
	payloadScheduleDay    = "sch_day_%d"
	payloadBackToMenu     = "backToMenu"
	payloadStudentRole    = "student"
	payloadAdminRole      = "admin"
	payloadTeacherRole    = "teacher"

	payloadBackToRoleSelection = "backToRoleSelection"
)

func GetSuperUserKeyboard(api *maxbot.Api) *maxbot.Keyboard {
	keyboard := api.Messages.NewKeyboardBuilder()
	keyboard.AddRow().AddCallback(btnAdmin, schemes.NEGATIVE, payloadAdminRole)
	keyboard.AddRow().AddCallback(btnTeacher, schemes.NEGATIVE, payloadTeacherRole)
	keyboard.AddRow().AddCallback(btnStudent, schemes.NEGATIVE, payloadStudentRole)
	return keyboard
}

func GetAdminKeyboard(api *maxbot.Api, isSuperUser bool) *maxbot.Keyboard {
	keyboard := api.Messages.NewKeyboardBuilder()
	keyboard.AddRow().AddCallback(btnUploadStudents, schemes.NEGATIVE, payloadUploadStudents)
	keyboard.AddRow().AddCallback(btnUploadTeachers, schemes.NEGATIVE, payloadUploadTeachers)
	keyboard.AddRow().AddCallback(btnUploadSchedule, schemes.NEGATIVE, payloadUploadSchedule)
	if isSuperUser {
		keyboard.AddRow().AddCallback(btnBackToRoleSelection, schemes.DEFAULT, payloadBackToRoleSelection)
	}
	return keyboard
}

func GetTeacherKeyboard(api *maxbot.Api, isSuperUser bool) *maxbot.Keyboard {
	keyboard := api.Messages.NewKeyboardBuilder()
	keyboard.AddRow().AddCallback(btnShowSchedule, schemes.NEGATIVE, payloadShowSchedule)
	keyboard.AddRow().AddCallback(btnMarkScore, schemes.NEGATIVE, payloadMarkGrade)
	keyboard.AddRow().AddCallback(btnMarkAttendance, schemes.NEGATIVE, payloadMarkAttendance)
	if isSuperUser {
		keyboard.AddRow().AddCallback(btnBackToRoleSelection, schemes.DEFAULT, payloadBackToRoleSelection)
	}
	return keyboard
}

func GetStudentKeyboard(api *maxbot.Api, isSuperUser bool) *maxbot.Keyboard {
	keyboard := api.Messages.NewKeyboardBuilder()
	keyboard.AddRow().AddCallback(btnShowSchedule, schemes.NEGATIVE, payloadShowSchedule)
	keyboard.AddRow().AddCallback(btnShowScore, schemes.NEGATIVE, payloadShowScore)
	keyboard.AddRow().AddCallback(btnShowAttendance, schemes.NEGATIVE, payloadShowAttendance)
	if isSuperUser {
		keyboard.AddRow().AddCallback(btnBackToRoleSelection, schemes.DEFAULT, payloadBackToRoleSelection)
	}
	return keyboard
}

func GetScheduleKeyboard(api *maxbot.Api, prev, next int16, isSuperUser bool) *maxbot.Keyboard {
	keyboard := api.Messages.NewKeyboardBuilder()
	keyboard.AddRow().
		AddCallback(btnPrev, schemes.NEGATIVE, fmt.Sprintf(payloadScheduleDay, prev)).
		AddCallback(btnNext, schemes.NEGATIVE, fmt.Sprintf(payloadScheduleDay, next))
	keyboard.AddRow().AddCallback(btnBackToMenu, schemes.DEFAULT, payloadBackToMenu)
	if isSuperUser {
		keyboard.AddRow().AddCallback(btnBackToRoleSelection, schemes.DEFAULT, payloadBackToRoleSelection)
	}
	return keyboard
}

func GetStudentsPaginationKeyboard(api *maxbot.Api, subjectID, groupID int64, currentPage, totalPages int, students []database.User, isSuperUser bool) *maxbot.Keyboard {
	keyboard := api.Messages.NewKeyboardBuilder()

	for _, student := range students {
		payload := fmt.Sprintf("grade_stud_%d_%d_%d", subjectID, groupID, student.UserID)
		keyboard.AddRow().AddCallback(student.Name, schemes.DEFAULT, payload)
	}

	if totalPages > 1 {
		row := keyboard.AddRow()

		if currentPage > 0 {
			prevPage := currentPage - 1
			payload := fmt.Sprintf("grade_stud_page_%d_%d_%d_%d", subjectID, groupID, prevPage, totalPages)
			row.AddCallback(btnPrev, schemes.NEGATIVE, payload)
		}

		if currentPage < totalPages-1 {
			nextPage := currentPage + 1
			payload := fmt.Sprintf("grade_stud_page_%d_%d_%d_%d", subjectID, groupID, nextPage, totalPages)
			row.AddCallback(btnNext, schemes.NEGATIVE, payload)
		}
	}

	keyboard.AddRow().AddCallback(btnBackToMenu, schemes.DEFAULT, payloadBackToMenu)
	if isSuperUser {
		keyboard.AddRow().AddCallback(btnBackToRoleSelection, schemes.DEFAULT, payloadBackToRoleSelection)
	}
	return keyboard
}
