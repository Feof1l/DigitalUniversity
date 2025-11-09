package maxAPI

import (
	"fmt"

	maxbot "github.com/max-messenger/max-bot-api-client-go"
	"github.com/max-messenger/max-bot-api-client-go/schemes"
)

func GetAdminKeyboard(api *maxbot.Api) *maxbot.Keyboard {
	keyboard := api.Messages.NewKeyboardBuilder()
	keyboard.AddRow().AddCallback("Загрузить файл со студентами", schemes.NEGATIVE, "uploadStudents")
	keyboard.AddRow().AddCallback("Загрузить файл с преподавателями", schemes.NEGATIVE, "uploadTeachers")
	keyboard.AddRow().AddCallback("Загрузить файл с расписанием", schemes.NEGATIVE, "uploadSchedule")
	return keyboard
}

func GetTeacherKeyboard(api *maxbot.Api) *maxbot.Keyboard {
	keyboard := api.Messages.NewKeyboardBuilder()
	keyboard.AddRow().AddCallback("Показать расписание", schemes.NEGATIVE, "showSchedule")
	keyboard.AddRow().AddCallback("Отметить посещаемость", schemes.NEGATIVE, "markAttendance")
	keyboard.AddRow().AddCallback("Поставить оценку", schemes.NEGATIVE, "markScore")
	return keyboard
}

func GetStudentKeyboard(api *maxbot.Api) *maxbot.Keyboard {
	keyboard := api.Messages.NewKeyboardBuilder()
	keyboard.AddRow().AddCallback("Показать расписание", schemes.NEGATIVE, "showSchedule")
	keyboard.AddRow().AddCallback("Посмотреть оценки", schemes.NEGATIVE, "showScore")
	return keyboard
}

func GetScheduleKeyboard(api *maxbot.Api, prev, next int16) *maxbot.Keyboard {
	keyboard := api.Messages.NewKeyboardBuilder()
	keyboard.AddRow().
		AddCallback("← Назад", schemes.NEGATIVE, fmt.Sprintf("sch_day_%d", prev)).
		AddCallback("Вперёд →", schemes.NEGATIVE, fmt.Sprintf("sch_day_%d", next))

	return keyboard
}
