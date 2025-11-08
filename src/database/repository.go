package database

import "github.com/jmoiron/sqlx"

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
    err := tx.Get(&groupID, `
        INSERT INTO groups (group_name)
        VALUES ($1)
        ON CONFLICT (group_name) DO UPDATE SET group_name = EXCLUDED.group_name
        RETURNING group_id`, groupName)
    return groupID, err
}

func (r *GroupRepository) GetGroupIDByName(tx *sqlx.Tx, groupName string) (int64, error) {
    var groupID int64
    err := tx.Get(&groupID, `SELECT group_id FROM groups WHERE group_name = $1`, groupName)
    return groupID, err
}

type UserRepository struct {
    db *sqlx.DB
}

func NewUserRepository(db *sqlx.DB) *UserRepository {
    return &UserRepository{db: db}
}

func (r *UserRepository) CreateOrUpdateStudent(tx *sqlx.Tx, userMaxID int64, firstName, lastName string, roleID, groupID int64) error {
    fullName := firstName + " " + lastName
    _, err := tx.Exec(`
        INSERT INTO users (name, usermax_id, first_name, last_name, role_id, group_id)
        VALUES ($1, $2, $3, $4, $5, $6)
        ON CONFLICT (usermax_id) DO UPDATE
        SET first_name = EXCLUDED.first_name,
            last_name = EXCLUDED.last_name,
            group_id = EXCLUDED.group_id`,
        fullName, userMaxID, firstName, lastName, roleID, groupID)
    return err
}

func (r *UserRepository) CreateOrUpdateTeacher(tx *sqlx.Tx, userMaxID int64, firstName, lastName string, roleID int64) error {
    fullName := firstName + " " + lastName
    _, err := tx.Exec(`
        INSERT INTO users (name, usermax_id, first_name, last_name, role_id, group_id)
        VALUES ($1, $2, $3, $4, $5, NULL)
        ON CONFLICT (usermax_id) DO UPDATE
        SET first_name = EXCLUDED.first_name,
            last_name = EXCLUDED.last_name`,
        fullName, userMaxID, firstName, lastName, roleID)
    return err
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

type ScheduleRepository struct {
    db *sqlx.DB
}

func NewScheduleRepository(db *sqlx.DB) *ScheduleRepository {
    return &ScheduleRepository{db: db}
}

func (r *ScheduleRepository) CreateSchedule(tx *sqlx.Tx, weekday int, startTime, endTime string, subjectID, teacherID, groupID, lessonTypeID int64) error {
    _, err := tx.Exec(`
        INSERT INTO schedule (weekday, start_time, end_time, subject_id, teacher_id, group_id, lesson_type_id)
        VALUES ($1, $2, $3, $4, $5, $6, $7)`,
        weekday, startTime, endTime, subjectID, teacherID, groupID, lessonTypeID)
    return err
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
    if err != nil {
        return "", err
    }
    return roleName, nil
}
