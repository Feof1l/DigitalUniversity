package database

import (
	"time"
)

type Role struct {
	RoleID   int64  `db:"role_id" json:"role_id"`
	RoleName string `db:"role_name" json:"role_name"`
}

type Group struct {
	GroupID   int64  `db:"group_id" json:"group_id"`
	GroupName string `db:"group_name" json:"group_name"`
}

type User struct {
	UserID    int64  `db:"user_id" json:"user_id"`
	Name      string `db:"name" json:"name"`
	UserMaxID int64  `db:"usermax_id" json:"usermax_id"`
	FirstName string `db:"first_name" json:"first_name"`
	LastName  string `db:"last_name" json:"last_name"`
	RoleID    int64  `db:"role_id" json:"role_id"`
	GroupID   *int64 `db:"group_id" json:"group_id"`
}

type Subject struct {
	SubjectID   int64  `db:"subject_id" json:"subject_id"`
	SubjectName string `db:"subject_name" json:"subject_name"`
	TeacherID   int64  `db:"teacher_id" json:"teacher_id"`
}

type GroupSubject struct {
	GroupID   int64 `db:"group_id" json:"group_id"`
	SubjectID int64 `db:"subject_id" json:"subject_id"`
}

type LessonType struct {
	LessonTypeID int64  `db:"lesson_type_id" json:"lesson_type_id"`
	TypeName     string `db:"type_name" json:"type_name"`
}

type Schedule struct {
	ScheduleID   int64     `db:"schedule_id" json:"schedule_id"`
	Weekday      int16     `db:"weekday" json:"weekday"`
	StartTime    time.Time `db:"start_time" json:"start_time"`
	EndTime      time.Time `db:"end_time" json:"end_time"`
	ClassRoom    string    `db:"class_room" json:"class_room"`
	SubjectID    int64     `db:"subject_id" json:"subject_id"`
	TeacherID    int64     `db:"teacher_id" json:"teacher_id"`
	GroupID      int64     `db:"group_id" json:"group_id"`
	LessonTypeID int64     `db:"lesson_type_id" json:"lesson_type_id"`
}

type Grade struct {
	GradeID    int64     `db:"grade_id" json:"grade_id"`
	StudentID  int64     `db:"student_id" json:"student_id"`
	TeacherID  int64     `db:"teacher_id" json:"teacher_id"`
	SubjectID  int64     `db:"subject_id" json:"subject_id"`
	ScheduleID int64     `db:"schedule_id" json:"schedule_id"`
	GradeValue int       `db:"grade_value" json:"grade_value"`
	GradeDate  time.Time `db:"grade_date" json:"grade_date"`
}

type Attendance struct {
	AttendanceID int64     `db:"attendance_id" json:"attendance_id"`
	StudentID    int64     `db:"student_id" json:"student_id"`
	ScheduleID   int64     `db:"schedule_id" json:"schedule_id"`
	Attended     bool      `db:"attended" json:"attended"`
	MarkTime     time.Time `db:"mark_time" json:"mark_time"`
}
