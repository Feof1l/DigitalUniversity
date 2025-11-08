package maxAPI

import (
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
