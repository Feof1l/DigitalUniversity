package maxAPI

import (
	"context"
	"digitalUniversity/database"
	"fmt"
)

func (b *Bot) sendSubjectsForTeacher(ctx context.Context, maxUserID int64) error {
	userRole, err := b.getUserRole(maxUserID)
	if err != nil {
		b.logger.Errorf("Failed to get user role: %v", err)
		return err
	}

	var entries []database.Schedule

	if userRole != "teacher" {
		return fmt.Errorf("у вас недостаточно прав")
	}

	teacherID, err := b.userRepo.GetUserIDByMaxID(maxUserID)
	if err != nil {
		b.logger.Errorf("Failed to get teacher ID: %v", err)
		return err
	}
	entries, err = b.scheduleRepo.GetScheduleByTeacher(teacherID)
	if err != nil {
		b.logger.Errorf("Failed to get schedule for teacher %d : %v", teacherID, err)
		return err
	}

	subjects := []string{}

	for _, entry := range entries {
		name, err := b.subjectRepo.GetSubjectName(entry.SubjectID)
		if err != nil {
			b.logger.Errorf("Failed to get subject name: %v", err)
		}

		subjects = append(subjects, name)
	}

	b.sendKeyboard(ctx, GetComplexKeyboard(b.MaxAPI, subjects), maxUserID, subjectsChooseMessage)

	return nil
}
