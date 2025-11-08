package maxAPI

import (
	"context"

	maxbot "github.com/max-messenger/max-bot-api-client-go"
	"github.com/max-messenger/max-bot-api-client-go/schemes"
)

func GetKeyboard(api *maxbot.Api, ctx context.Context) *maxbot.Keyboard {
	keyboard := api.Messages.NewKeyboardBuilder()
	keyboard.
		AddRow().
		AddCallback("Загрузить файл со студентами", schemes.NEGATIVE, "uploadStudents")
	keyboard.
		AddRow().
		AddCallback("Загрузить файл с преподавателями", schemes.NEGATIVE, "uploadTeachers")
	keyboard.
		AddRow().
		AddCallback("Загрузить файл с расписанием", schemes.NEGATIVE, "uploadSchedule")

	return keyboard
}
