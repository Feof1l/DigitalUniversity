package database

import (
	"github.com/jmoiron/sqlx"
)

type RoleRepository struct {
	db *sqlx.DB
}

func NewRoleRepository(db *sqlx.DB) *RoleRepository {
	return &RoleRepository{db: db}
}

func (r *RoleRepository) GetRoleIDByName(tx *sqlx.Tx, roleName string) (int64, error) {
	var roleID int64
	err := tx.Get(&roleID, `SELECT role_id FROM roles WHERE role_name = $1`, roleName)
	return roleID, err
}

type GroupRepository struct {
	db *sqlx.DB
}

func NewGroupRepository(db *sqlx.DB) *GroupRepository {
	return &GroupRepository{db: db}
}

func (r *GroupRepository) CreateOrGetGroup(tx *sqlx.Tx, groupName string) (int64, error) {
	var groupID int64
	_, err := tx.Exec(`
        INSERT INTO groups (group_name)
        VALUES ($1)
        ON CONFLICT (group_name) DO NOTHING`, groupName)
	if err != nil {
		return 0, err
	}
	err = tx.Get(&groupID, `SELECT group_id FROM groups WHERE group_name = $1`, groupName)
	return groupID, err
}

func (r *GroupRepository) GetGroupIDByName(tx *sqlx.Tx, groupName string) (int64, error) {
	var groupID int64
	err := tx.Get(&groupID, `SELECT group_id FROM groups WHERE group_name = $1`, groupName)
	return groupID, err
}

func (r *GroupRepository) GetGroupName(groupID int64) (string, error) {
	var groupName string
	err := r.db.Get(&groupName, `SELECT group_name FROM groups WHERE group_id = $1`, groupID)
	return groupName, err
}

type UserRepository struct {
	db *sqlx.DB
}

func NewUserRepository(db *sqlx.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) GetUserMaxIDByID(userID int64) (int64, error) {
	var userMaxID int64
	err := r.db.Get(&userMaxID, `SELECT usermax_id FROM users WHERE user_id = $1`, userID)
	return userMaxID, err
}

func (r *UserRepository) CreateOrUpdateStudent(tx *sqlx.Tx, userMaxID int64, firstName, lastName string, roleID, groupID int64) error {
	fullName := firstName + " " + lastName

	result, err := tx.Exec(`
		UPDATE users
		SET usermax_id = $1, name = $2
		WHERE first_name = $3 AND last_name = $4 AND role_id = $5 AND group_id = $6`,
		userMaxID, fullName, firstName, lastName, roleID, groupID)
	if err != nil {
		return err
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected > 0 {
		return nil
	}

	_, err = tx.Exec(`
		INSERT INTO users (name, usermax_id, first_name, last_name, role_id, group_id)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (usermax_id) DO UPDATE
		SET first_name = EXCLUDED.first_name,
		    last_name = EXCLUDED.last_name,
		    name = EXCLUDED.name,
		    group_id = EXCLUDED.group_id`,
		fullName, userMaxID, firstName, lastName, roleID, groupID)

	return err
}

func (r *UserRepository) UpdateUserRole(tx *sqlx.Tx, userMaxID int64, roleID int64) error {

	_, err := tx.Exec(`
		UPDATE users
		SET role_id = $1
		WHERE usermax_id = $2`,
		roleID, userMaxID)
	if err != nil {
		return err
	}

	// rowsAffected, _ := result.RowsAffected()
	// if rowsAffected > 0 {
	// 	return nil
	// }

	return err
}

func (r *UserRepository) CreateOrUpdateTeacher(tx *sqlx.Tx, userMaxID int64, firstName, lastName string, roleID int64) error {
	fullName := firstName + " " + lastName

	result, err := tx.Exec(`
		UPDATE users
		SET usermax_id = $1, name = $2
		WHERE first_name = $3 AND last_name = $4 AND role_id = $5 AND group_id IS NULL`,
		userMaxID, fullName, firstName, lastName, roleID)
	if err != nil {
		return err
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected > 0 {
		return nil
	}

	_, err = tx.Exec(`
		INSERT INTO users (name, usermax_id, first_name, last_name, role_id, group_id)
		VALUES ($1, $2, $3, $4, $5, NULL)
		ON CONFLICT (usermax_id) DO UPDATE
		SET first_name = EXCLUDED.first_name,
		    last_name = EXCLUDED.last_name,
		    name = EXCLUDED.name`,
		fullName, userMaxID, firstName, lastName, roleID)

	return err
}

func (r *UserRepository) CreateOrGetTeacher(tx *sqlx.Tx, firstName, lastName string, teacherRoleID int64) (int64, error) {
	var teacherID int64

	err := tx.Get(&teacherID,
		`SELECT user_id FROM users
		WHERE first_name = $1 AND last_name = $2 AND role_id = $3 AND group_id IS NULL`,
		firstName, lastName, teacherRoleID)

	if err == nil {
		return teacherID, nil
	}

	row := tx.QueryRow(`
        INSERT INTO users (name, first_name, last_name, role_id, group_id)
        VALUES ($1, $2, $3, $4, NULL)
        RETURNING user_id`,
		firstName+" "+lastName, firstName, lastName, teacherRoleID)
	err = row.Scan(&teacherID)
	return teacherID, err
}

func (r *UserRepository) GetTeacherIDByName(tx *sqlx.Tx, lastName, firstName string) (int64, error) {
	var teacherID int64
	err := tx.Get(&teacherID, `
        SELECT user_id FROM users
        WHERE last_name = $1 AND first_name = $2
        AND role_id = (SELECT role_id FROM roles WHERE role_name = 'teacher')`,
		lastName, firstName)
	return teacherID, err
}

func (r *UserRepository) GetUserByMaxID(userMaxID int64) (*User, error) {
	user := new(User)
	err := r.db.Get(user, `SELECT * FROM users WHERE usermax_id = $1`, userMaxID)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (r *UserRepository) GetUserRole(userMaxID int64) (string, error) {
	var roleName string
	err := r.db.Get(&roleName, `
        SELECT r.role_name
        FROM users u
        JOIN roles r ON u.role_id = r.role_id
        WHERE u.usermax_id = $1`, userMaxID)
	return roleName, err
}

func (r *UserRepository) GetTeacherName(teacherID int64) (string, error) {
	var teacherName string
	err := r.db.Get(&teacherName, `SELECT name FROM users WHERE user_id = $1`, teacherID)
	return teacherName, err
}

func (r *UserRepository) GetUserIDByMaxID(userMaxID int64) (int64, error) {
	var userID int64
	err := r.db.Get(&userID, `SELECT user_id FROM users WHERE usermax_id = $1`, userMaxID)
	return userID, err
}

func (r *UserRepository) GetStudentGroupID(userID int64) (int64, error) {
	var groupID int64
	err := r.db.Get(&groupID, `
        SELECT group_id
        FROM users
        WHERE user_id = $1 AND group_id IS NOT NULL`, userID)
	return groupID, err
}

type SubjectRepository struct {
	db *sqlx.DB
}

func NewSubjectRepository(db *sqlx.DB) *SubjectRepository {
	return &SubjectRepository{db: db}
}

func (r *SubjectRepository) CreateOrGetSubject(tx *sqlx.Tx, subjectName string, teacherID int64) (int64, error) {
	var subjectID int64
	err := tx.Get(&subjectID, `
        INSERT INTO subjects (subject_name, teacher_id)
        VALUES ($1, $2)
        ON CONFLICT DO NOTHING
        RETURNING subject_id`, subjectName, teacherID)
	if err != nil {
		err = tx.Get(&subjectID, `SELECT subject_id FROM subjects WHERE subject_name = $1`, subjectName)
	}
	return subjectID, err
}

func (r *SubjectRepository) LinkGroupToSubject(tx *sqlx.Tx, groupID, subjectID int64) error {
	_, err := tx.Exec(`
        INSERT INTO groups_subjects (group_id, subject_id)
        VALUES ($1, $2)
        ON CONFLICT DO NOTHING`, groupID, subjectID)
	return err
}

func (r *SubjectRepository) GetSubjectName(subjectID int64) (string, error) {
	var subjectName string
	err := r.db.Get(&subjectName, `SELECT subject_name FROM subjects WHERE subject_id = $1`, subjectID)
	return subjectName, err
}

type LessonTypeRepository struct {
	db *sqlx.DB
}

func NewLessonTypeRepository(db *sqlx.DB) *LessonTypeRepository {
	return &LessonTypeRepository{db: db}
}

func (r *LessonTypeRepository) CreateOrGetLessonType(tx *sqlx.Tx, typeName string) (int64, error) {
	var lessonTypeID int64
	err := tx.Get(&lessonTypeID, `
        INSERT INTO lesson_types (type_name)
        VALUES ($1)
        ON CONFLICT (type_name) DO UPDATE SET type_name = EXCLUDED.type_name
        RETURNING lesson_type_id`, typeName)
	return lessonTypeID, err
}

func (r *LessonTypeRepository) GetLessonTypeName(lessonTypeID int64) (string, error) {
	var lessonTypeName string
	err := r.db.Get(&lessonTypeName, `SELECT type_name FROM lesson_types WHERE lesson_type_id = $1`, lessonTypeID)
	return lessonTypeName, err
}

type ScheduleRepository struct {
	db *sqlx.DB
}

func NewScheduleRepository(db *sqlx.DB) *ScheduleRepository {
	return &ScheduleRepository{db: db}
}

func (r *ScheduleRepository) CreateSchedule(tx *sqlx.Tx, weekday int, startTime, endTime, classroom string, subjectID, teacherID, groupID, lessonTypeID int64) error {
	_, err := tx.Exec(`
        INSERT INTO schedule (weekday, start_time, end_time, class_room, subject_id, teacher_id, group_id, lesson_type_id)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		weekday, startTime, endTime, classroom, subjectID, teacherID, groupID, lessonTypeID)
	return err
}

func (r *ScheduleRepository) GetScheduleForDate(weekday int16) ([]Schedule, error) {
	var entries []Schedule
	query := `SELECT * FROM schedule WHERE weekday = $1 ORDER BY start_time`
	err := r.db.Select(&entries, query, weekday)
	return entries, err
}

func (r *ScheduleRepository) GetScheduleForDateByTeacher(weekday int16, teacherID int64) ([]Schedule, error) {
	var entries []Schedule
	query := `SELECT * FROM schedule WHERE weekday = $1 AND teacher_id = $2 ORDER BY start_time`
	err := r.db.Select(&entries, query, weekday, teacherID)
	return entries, err
}

func (r *ScheduleRepository) GetScheduleForDateByGroup(weekday int16, groupID int64) ([]Schedule, error) {
	var entries []Schedule
	query := `SELECT * FROM schedule WHERE weekday = $1 AND group_id = $2 ORDER BY start_time`
	err := r.db.Select(&entries, query, weekday, groupID)
	return entries, err
}

type GradeRepository struct {
	db *sqlx.DB
}

func NewGradeRepository(db *sqlx.DB) *GradeRepository {
	return &GradeRepository{db: db}
}

func (r *GradeRepository) GetSubjectsByTeacher(teacherID int64) ([]Subject, error) {
	var subjects []Subject
	query := `SELECT * FROM subjects WHERE teacher_id = $1 ORDER BY subject_name`
	err := r.db.Select(&subjects, query, teacherID)
	return subjects, err
}

func (r *GradeRepository) GetSubjects() ([]Subject, error) {
	var subjects []Subject
	query := `SELECT * FROM subjects ORDER BY subject_name`
	err := r.db.Select(&subjects, query)
	return subjects, err
}

func (r *GradeRepository) GetSubjectsByStudentGroup(groupID int64) ([]Subject, error) {
	var subjects []Subject
	query := `
        SELECT DISTINCT s.subject_id, s.subject_name, s.teacher_id
        FROM subjects s
        JOIN groups_subjects gs ON s.subject_id = gs.subject_id
        WHERE gs.group_id = $1
        ORDER BY s.subject_name`
	err := r.db.Select(&subjects, query, groupID)
	return subjects, err
}

func (r *GradeRepository) GetGradesByStudentAndSubject(studentID, subjectID int64) ([]Grade, error) {
	var grades []Grade
	query := `SELECT * FROM grades WHERE student_id = $1 AND subject_id = $2 ORDER BY grade_date DESC`
	err := r.db.Select(&grades, query, studentID, subjectID)
	return grades, err
}

func (r *GradeRepository) GetGradesBySubject(subjectID int64) ([]Grade, error) {
	var grades []Grade
	query := `SELECT * FROM grades WHERE subject_id = $1 ORDER BY grade_date DESC`
	err := r.db.Select(&grades, query, subjectID)
	return grades, err
}

func (r *GradeRepository) GetGroupsBySubjectAndTeacher(subjectID, teacherID int64) ([]Group, error) {
	var groups []Group
	query := `
        SELECT DISTINCT g.group_id, g.group_name
        FROM groups g
        JOIN groups_subjects gs ON g.group_id = gs.group_id
        JOIN subjects s ON gs.subject_id = s.subject_id
        WHERE s.subject_id = $1 AND s.teacher_id = $2
        ORDER BY g.group_name`
	err := r.db.Select(&groups, query, subjectID, teacherID)
	return groups, err
}

func (r *GradeRepository) GetGroupsBySubject(subjectID int64) ([]Group, error) {
	var groups []Group
	query := `
        SELECT DISTINCT g.group_id, g.group_name
        FROM groups g
        JOIN groups_subjects gs ON g.group_id = gs.group_id
        JOIN subjects s ON gs.subject_id = s.subject_id
        WHERE s.subject_id = $1 
        ORDER BY g.group_name`
	err := r.db.Select(&groups, query, subjectID)
	return groups, err
}

func (r *GradeRepository) GetStudentsByGroup(groupID int64) ([]User, error) {
	var students []User
	query := `
        SELECT * FROM users
        WHERE group_id = $1 AND role_id = (SELECT role_id FROM roles WHERE role_name = 'student')
        ORDER BY last_name, first_name`
	err := r.db.Select(&students, query, groupID)
	return students, err
}

func (r *GradeRepository) GetScheduleBySubjectAndGroup(subjectID, groupID int64) ([]Schedule, error) {
	var schedules []Schedule
	query := `
        SELECT * FROM schedule
        WHERE subject_id = $1 AND group_id = $2
        ORDER BY weekday, start_time`
	err := r.db.Select(&schedules, query, subjectID, groupID)
	return schedules, err
}

func (r *GradeRepository) CreateGrade(studentID, teacherID, subjectID, scheduleID int64, gradeValue int) error {
	_, err := r.db.Exec(`
        INSERT INTO grades (student_id, teacher_id, subject_id, schedule_id, grade_value, grade_date)
        VALUES ($1, $2, $3, $4, $5, NOW())`,
		studentID, teacherID, subjectID, scheduleID, gradeValue)
	return err
}

func (r *GradeRepository) GetGradesByStudent(studentID int64) ([]Grade, error) {
	var grades []Grade
	query := `SELECT * FROM grades WHERE student_id = $1 ORDER BY grade_date DESC`
	err := r.db.Select(&grades, query, studentID)
	return grades, err
}

func (r *GradeRepository) GetStudentNameByID(studentID int64) (string, error) {
	var studentName string
	err := r.db.Get(&studentName, `SELECT name FROM users WHERE user_id = $1`, studentID)
	return studentName, err
}

func (r *GradeRepository) GetSubjectIDByScheduleID(scheduleID int64) (int64, error) {
	var subjectID int64
	err := r.db.Get(&subjectID, `SELECT subject_id FROM schedule WHERE schedule_id = $1`, scheduleID)
	return subjectID, err
}

type AttendanceRepository struct {
	db *sqlx.DB
}

func NewAttendanceRepository(db *sqlx.DB) *AttendanceRepository {
	return &AttendanceRepository{db: db}
}

func (r *AttendanceRepository) MarkAttendance(studentID, scheduleID int64, attended bool) error {
	var count int
	err := r.db.Get(&count, `
        SELECT COUNT(*) FROM attendance
        WHERE student_id = $1 AND schedule_id = $2`, studentID, scheduleID)
	if err != nil {
		return err
	}

	if count > 0 {
		_, err = r.db.Exec(`
            UPDATE attendance
            SET attended = $1, mark_time = NOW()
            WHERE student_id = $2 AND schedule_id = $3`,
			attended, studentID, scheduleID)
		return err
	}

	_, err = r.db.Exec(`
        INSERT INTO attendance (student_id, schedule_id, attended, mark_time)
        VALUES ($1, $2, $3, NOW())`,
		studentID, scheduleID, attended)
	return err
}

func (r *AttendanceRepository) GetAttendanceByStudentAndSubject(studentID, subjectID int64) ([]Attendance, error) {
	var attendance []Attendance
	query := `
        SELECT a.* FROM attendance a
        JOIN schedule s ON a.schedule_id = s.schedule_id
        WHERE a.student_id = $1 AND s.subject_id = $2
        ORDER BY a.mark_time DESC`
	err := r.db.Select(&attendance, query, studentID, subjectID)
	return attendance, err
}

func (r *AttendanceRepository) GetAttendanceBySubject(subjectID int64) ([]Attendance, error) {
	var attendance []Attendance
	query := `
        SELECT a.* FROM attendance a
        JOIN schedule s ON a.schedule_id = s.schedule_id
        WHERE s.subject_id = $1
        ORDER BY a.mark_time DESC`
	err := r.db.Select(&attendance, query, subjectID)
	return attendance, err
}

func (r *AttendanceRepository) GetMarkedStudentIDsBySchedule(scheduleID int64) ([]int64, error) {
	var studentIDs []int64
	query := `SELECT student_id FROM attendance WHERE schedule_id = $1`
	err := r.db.Select(&studentIDs, query, scheduleID)
	return studentIDs, err
}

func (r *AttendanceRepository) GetAttendanceRecordsBySchedule(scheduleID int64) ([]Attendance, error) {
	var records []Attendance
	query := `SELECT * FROM attendance WHERE schedule_id = $1`
	err := r.db.Select(&records, query, scheduleID)
	return records, err
}
