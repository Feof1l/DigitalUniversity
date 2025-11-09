package maxAPI

import (
	"fmt"

	maxbot "github.com/max-messenger/max-bot-api-client-go"
	"github.com/max-messenger/max-bot-api-client-go/schemes"
)

const (
	btnUploadStudents  = "Загрузить файл со студентами"
	btnUploadTeachers  = "Загрузить файл с преподавателями"
	btnUploadSchedule  = "Загрузить файл с расписанием"

	btnShowSchedule     = "Показать расписание"
	btnMarkAttendance   = "Отметить посещаемость"
	btnMarkScore        = "Поставить оценку"

	btnShowScore = "Посмотреть оценки"

	btnPrev = "← Назад"
	btnNext = "Вперёд →"

	payloadUploadStudents  = "uploadStudents"
	payloadUploadTeachers  = "uploadTeachers"
	payloadUploadSchedule  = "uploadSchedule"
	payloadShowSchedule    = "showSchedule"
	payloadMarkAttendance  = "markAttendance"
	payloadMarkScore       = "markScore"
	payloadShowScore       = "showScore"
	payloadScheduleDay     = "sch_day_%d"
)

func GetAdminKeyboard(api *maxbot.Api) *maxbot.Keyboard {
	keyboard := api.Messages.NewKeyboardBuilder()
	keyboard.AddRow().AddCallback(btnUploadStudents, schemes.NEGATIVE, payloadUploadStudents)
	keyboard.AddRow().AddCallback(btnUploadTeachers, schemes.NEGATIVE, payloadUploadTeachers)
	keyboard.AddRow().AddCallback(btnUploadSchedule, schemes.NEGATIVE, payloadUploadSchedule)
	return keyboard
}

func GetTeacherKeyboard(api *maxbot.Api) *maxbot.Keyboard {
	keyboard := api.Messages.NewKeyboardBuilder()
	keyboard.AddRow().AddCallback(btnShowSchedule, schemes.NEGATIVE, payloadShowSchedule)
	keyboard.AddRow().AddCallback(btnMarkAttendance, schemes.NEGATIVE, payloadMarkAttendance)
	keyboard.AddRow().AddCallback(btnMarkScore, schemes.NEGATIVE, payloadMarkScore)
	return keyboard
}

func GetStudentKeyboard(api *maxbot.Api) *maxbot.Keyboard {
	keyboard := api.Messages.NewKeyboardBuilder()
	keyboard.AddRow().AddCallback(btnShowSchedule, schemes.NEGATIVE, payloadShowSchedule)
	keyboard.AddRow().AddCallback(btnShowScore, schemes.NEGATIVE, payloadShowScore)
	return keyboard
}

func GetScheduleKeyboard(api *maxbot.Api, prev, next int16) *maxbot.Keyboard {
	keyboard := api.Messages.NewKeyboardBuilder()
	keyboard.AddRow().
		AddCallback(btnPrev, schemes.NEGATIVE, fmt.Sprintf(payloadScheduleDay, prev)).
		AddCallback(btnNext, schemes.NEGATIVE, fmt.Sprintf(payloadScheduleDay, next))
	return keyboard
}
