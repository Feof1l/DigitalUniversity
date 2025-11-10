package maxAPI

import (
	"context"
	"digitalUniversity/database"
	"fmt"
)

func (b *Bot) sendGroupsForTeacher(ctx context.Context, maxUserID int64) error {
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

	groups := []string{}

	for _, entry := range entries {
		name, err := b.groupRepo.GetGroupName(entry.GroupID)
		if err != nil {
			b.logger.Errorf("Failed to get subject name: %v", err)
		}

		groups = append(groups, name)
	}

	b.sendKeyboard(ctx, GetComplexKeyboard(b.MaxAPI, groups), maxUserID, groupsChooseMessage)

	return nil
}

func (b *Bot) sendStudentsFromGroup(ctx context.Context, groupName string, maxUserID int64) error {
	userRole, err := b.getUserRole(maxUserID)
	if err != nil {
		b.logger.Errorf("Failed to get user role: %v", err)
		return err
	}

	//var entries []database.Schedule

	if userRole != "teacher" {
		return fmt.Errorf("у вас недостаточно прав")
	}

	// teacherID, err := b.userRepo.GetUserIDByMaxID(maxUserID)
	// if err != nil {
	// 	b.logger.Errorf("Failed to get teacher ID: %v", err)
	// 	return err
	// }
	tx, err := b.db.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	groupID, err := b.groupRepo.GetGroupIDByName(tx, groupName)
	if err != nil {
		b.logger.Errorf("Failed to get group id: %v", err)
		return err
	}

	if userRole != "teacher" {
		return fmt.Errorf("у вас недостаточно прав")
	}

	students, err := b.userRepo.GetUsersByGroupID(groupID)
	if err != nil {
		b.logger.Errorf("Failed to get students : %v", err)
		return err
	}

	studentNames := []string{}

	for _, student := range students {
		studentNames = append(studentNames, student.Name)
	}

	b.sendKeyboard(ctx, GetComplexKeyboard(b.MaxAPI, studentNames), maxUserID, studentsChooseMessage)

	return nil
}
