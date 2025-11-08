package database

import (
	"encoding/csv"
	"os"
	"strconv"

	"github.com/jmoiron/sqlx"
)

type CSVImporter struct {
    roleRepo       *RoleRepository
    groupRepo      *GroupRepository
    userRepo       *UserRepository
    subjectRepo    *SubjectRepository
    lessonTypeRepo *LessonTypeRepository
    scheduleRepo   *ScheduleRepository
    db             *sqlx.DB
}

func NewCSVImporter(db *sqlx.DB) *CSVImporter {
    return &CSVImporter{
        roleRepo:       NewRoleRepository(db),
        groupRepo:      NewGroupRepository(db),
        userRepo:       NewUserRepository(db),
        subjectRepo:    NewSubjectRepository(db),
        lessonTypeRepo: NewLessonTypeRepository(db),
        scheduleRepo:   NewScheduleRepository(db),
        db:             db,
    }
}

func (imp *CSVImporter) ImportStudents(filePath string) error {
    records, err := readCSV(filePath)
    if err != nil {
        return err
    }

    tx, err := imp.db.Beginx()
    if err != nil {
        return err
    }
    defer tx.Rollback()

    studentRoleID, err := imp.roleRepo.GetRoleIDByName(tx, "student")
    if err != nil {
        return err
    }

    for i, record := range records {
        if i == 0 {
            continue
        }

        userMaxID, _ := strconv.ParseInt(record[0], 10, 64)
        lastName := record[1]
        firstName := record[2]
        groupName := record[3]

        groupID, err := imp.groupRepo.CreateOrGetGroup(tx, groupName)
        if err != nil {
            return err
        }

        err = imp.userRepo.CreateOrUpdateStudent(tx, userMaxID, firstName, lastName, studentRoleID, groupID)
        if err != nil {
            return err
        }
    }

    return tx.Commit()
}

func (imp *CSVImporter) ImportTeachers(filePath string) error {
    records, err := readCSV(filePath)
    if err != nil {
        return err
    }

    tx, err := imp.db.Beginx()
    if err != nil {
        return err
    }
    defer tx.Rollback()

    teacherRoleID, err := imp.roleRepo.GetRoleIDByName(tx, "teacher")
    if err != nil {
        return err
    }

    for i, record := range records {
        if i == 0 {
            continue
        }

        userMaxID, _ := strconv.ParseInt(record[0], 10, 64)
        lastName := record[1]
        firstName := record[2]

        err = imp.userRepo.CreateOrUpdateTeacher(tx, userMaxID, firstName, lastName, teacherRoleID)
        if err != nil {
            return err
        }
    }

    return tx.Commit()
}

func (imp *CSVImporter) ImportSchedule(filePath string) error {
    records, err := readCSV(filePath)
    if err != nil {
        return err
    }

    tx, err := imp.db.Beginx()
    if err != nil {
        return err
    }
    defer tx.Rollback()

    for i, record := range records {
        if i == 0 {
            continue
        }

        subjectName := record[0]
        typeName := record[1]
        groupName := record[3]
        teacherLastName := record[4]
        teacherFirstName := record[5]
        weekday, _ := strconv.Atoi(record[6])
        startTime := record[7]
        endTime := record[8]

        lessonTypeID, err := imp.lessonTypeRepo.CreateOrGetLessonType(tx, typeName)
        if err != nil {
            return err
        }

        teacherID, err := imp.userRepo.GetTeacherIDByName(tx, teacherLastName, teacherFirstName)
        if err != nil {
            return err
        }

        subjectID, err := imp.subjectRepo.CreateOrGetSubject(tx, subjectName, teacherID)
        if err != nil {
            return err
        }

        groupID, err := imp.groupRepo.GetGroupIDByName(tx, groupName)
        if err != nil {
            return err
        }

        err = imp.scheduleRepo.CreateSchedule(tx, weekday, startTime, endTime, subjectID, teacherID, groupID, lessonTypeID)
        if err != nil {
            return err
        }

        err = imp.subjectRepo.LinkGroupToSubject(tx, groupID, subjectID)
        if err != nil {
            return err
        }
    }

    return tx.Commit()
}

func readCSV(filePath string) ([][]string, error) {
    file, err := os.Open(filePath)
    if err != nil {
        return nil, err
    }
    defer file.Close()

    reader := csv.NewReader(file)
    return reader.ReadAll()
}
