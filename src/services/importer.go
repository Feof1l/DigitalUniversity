package services

import (
	"encoding/csv"
	"os"
	"strconv"

	"github.com/jmoiron/sqlx"

	"digitalUniversity/database"
)

type CSVImporter struct {
	roleRepo       *database.RoleRepository
	groupRepo      *database.GroupRepository
	userRepo       *database.UserRepository
	subjectRepo    *database.SubjectRepository
	lessonTypeRepo *database.LessonTypeRepository
	scheduleRepo   *database.ScheduleRepository
	db             *sqlx.DB
}

func NewCSVImporter(db *sqlx.DB) *CSVImporter {
	return &CSVImporter{
		roleRepo:       database.NewRoleRepository(db),
		groupRepo:      database.NewGroupRepository(db),
		userRepo:       database.NewUserRepository(db),
		subjectRepo:    database.NewSubjectRepository(db),
		lessonTypeRepo: database.NewLessonTypeRepository(db),
		scheduleRepo:   database.NewScheduleRepository(db),
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

	for i := 1; i < len(records); i++ {
		record := records[i]

		userMaxID, err := strconv.ParseInt(record[0], 10, 64)
		if err != nil {
			return err
		}

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

	for i := 1; i < len(records); i++ {
		record := records[i]

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

	teacherRoleID, err := imp.roleRepo.GetRoleIDByName(tx, "teacher")
	if err != nil {
		return err
	}

	for i := 1; i < len(records); i++ {
		record := records[i]

		subjectName := record[0]
		typeName := record[1]
		classroom := record[2]
		groupName := record[3]
		teacherLastName := record[4]
		teacherFirstName := record[5]

		weekday, err := strconv.Atoi(record[6])
		if err != nil {
			return err
		}

		startTime := record[7]
		endTime := record[8]

		lessonTypeID, err := imp.lessonTypeRepo.CreateOrGetLessonType(tx, typeName)
		if err != nil {
			return err
		}

		teacherID, err := imp.userRepo.CreateOrGetTeacher(tx, teacherFirstName, teacherLastName, teacherRoleID)
		if err != nil {
			return err
		}

		subjectID, err := imp.subjectRepo.CreateOrGetSubject(tx, subjectName, teacherID)
		if err != nil {
			return err
		}

		groupID, err := imp.groupRepo.CreateOrGetGroup(tx, groupName)
		if err != nil {
			return err
		}

		err = imp.scheduleRepo.CreateSchedule(tx, weekday, startTime, endTime, classroom, subjectID, teacherID, groupID, lessonTypeID)
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
