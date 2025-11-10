package maxAPI

// func (b *Bot) sendAttendanceForTeacher(ctx context.Context, u *schemes.MessageCallbackUpdate, maxUserID int64) error {
// 	userRole, err := b.getUserRole(maxUserID)
// 	if err != nil {
// 		b.logger.Errorf("Failed to get user role: %v", err)
// 		return err
// 	}

// 	var entries []database.Schedule

// 	if userRole != "teacher" {
// 		return fmt.Errorf("у вас недостаточно прав")
// 	}

// 	teacherID, err := b.userRepo.GetUserIDByMaxID(maxUserID)
// 	if err != nil {
// 		b.logger.Errorf("Failed to get teacher ID: %v", err)
// 		return err
// 	}
// 	entries, err = b.scheduleRepo.GetScheduleByTeacher(teacherID)
// 	if err != nil {
// 		b.logger.Errorf("Failed to get schedule for teacher %d, weekday %d: %v", teacherID, weekday, err)
// 		return err
// 	}

// 	subjects := []string{}

// 	for _, entry := range entries {
// 		name, err := b.subjectRepo.GetSubjectName(entry.SubjectID)
// 		if err != nil {
// 			b.logger.Errorf("Failed to get subject name: %v", err)
// 		}

// 		subjects = append(subjects, name)
// 	}

// 	text := b.formatSchedule(entries, weekday)
// 	b.logger.Infof("Sending schedule for weekday %d to user %d", weekday, maxUserID)

// 	prevDay, nextDay := b.calculateNavigationDays(weekday)

// 	// b.logger.Infof("aaaaaaaaaaaaaaaaa %s: %v %v", b.scheduleMessageIDs[maxUserID], maxUserID, u.Message.Body.Mid)
// 	// if msgID, exists := b.scheduleMessageIDs[maxUserID]; exists {

// 	// 	mID, err := strconv.Atoi(msgID)
// 	// 	if err != nil {
// 	// 		b.logger.Errorf("failed to convert msgID %v", err)
// 	// 	}
// 	// 	err = b.MaxAPI.Messages.EditMessage(ctx, int64(mID), maxbot.NewMessage().
// 	// 		SetUser(maxUserID).
// 	// 		AddKeyboard(GetScheduleKeyboard(b.MaxAPI, prevDay, nextDay)).
// 	// 		SetText(text))

// 	// 	if err != nil {
// 	// 		b.logger.Errorf("Failed to edit message %s for user %d: %v", msgID, maxUserID, err)
// 	// 		delete(b.scheduleMessageIDs, maxUserID)
// 	// 		b.sendKeyboard(ctx, GetScheduleKeyboard(b.MaxAPI, prevDay, nextDay), maxUserID, text)

// 	// 		//return b.sendNewScheduleMessage(ctx, maxUserID, weekday, text, keyboard)
// 	// 	}
// 	// 	return nil
// 	// } else {
// 	// 	//delete(b.scheduleMessageIDs, maxUserID)

// 	// 	//b.sendKeyboard(ctx, GetScheduleKeyboard(b.MaxAPI, prevDay, nextDay), maxUserID, text)

// 	// 	id, err := b.MaxAPI.Messages.Send(ctx, maxbot.NewMessage().
// 	// 		SetUser(maxUserID).
// 	// 		AddKeyboard(GetScheduleKeyboard(b.MaxAPI, prevDay, nextDay)).
// 	// 		SetText(text))
// 	// 	b.logger.Errorf("dsdsds %s: %v %v %v", msgID, maxUserID, u.Message.Body.Mid, id)

// 	// 	if err != nil && err.Error() != "" {
// 	// 		b.logger.Errorf("Failed to send keyboard: %v", err)
// 	// 	}
// 	// 	ids:=u.Message.Body.

// 	// 	b.scheduleMessageIDs[maxUserID] = u.Message.Body.Mid
// 	// }

// 	b.sendKeyboard(ctx, GetScheduleKeyboard(b.MaxAPI, prevDay, nextDay), maxUserID, text)

// 	return nil
// }
