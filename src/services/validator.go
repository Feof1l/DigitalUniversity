package services

import (
	"encoding/csv"
	"fmt"
	"os"
)

type FileType string

const (
	FileTypeStudents FileType = "students"
	FileTypeTeachers FileType = "teachers"
	FileTypeSchedule FileType = "schedule"
)

type ValidationError struct {
	Message string
}

func (e *ValidationError) Error() string {
	return e.Message
}

func ValidateCSVStructure(filePath string, expectedType FileType) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return &ValidationError{Message: "Ошибка чтения CSV файла. Убедитесь что файл имеет правильный формат."}
	}

	if len(records) == 0 {
		return &ValidationError{Message: "Файл пустой. Отправьте файл с данными."}
	}

	if len(records) == 1 {
		return &ValidationError{Message: "Файл содержит только заголовки. Добавьте данные."}
	}

	header := records[0]

	switch expectedType {
	case FileTypeStudents:
		expectedHeaders := []string{"User_id", "Last_name", "First_name", "Study_group"}
		if !validateHeaders(header, expectedHeaders) {
			return &ValidationError{
				Message: fmt.Sprintf("Неверная структура файла студентов.\n\nОжидаются столбцы:\n%v\n\nПолучены:\n%v\n\nОтправьте правильный файл со студентами.",
					expectedHeaders, header),
			}
		}

	case FileTypeTeachers:
		expectedHeaders := []string{"User_id", "Last_name", "First_name"}
		if !validateHeaders(header, expectedHeaders) {
			return &ValidationError{
				Message: fmt.Sprintf("Неверная структура файла преподавателей.\n\nОжидаются столбцы:\n%v\n\nПолучены:\n%v\n\nОтправьте правильный файл с преподавателями.",
					expectedHeaders, header),
			}
		}

	case FileTypeSchedule:
		expectedHeaders := []string{"subject_name", "type_name", "classroom", "group_name", "teacher_last_name", "teacher_first_name", "weekday", "start_time", "end_time", "lesson_type_id"}
		if !validateHeaders(header, expectedHeaders) {
			return &ValidationError{
				Message: fmt.Sprintf("Неверная структура файла расписания.\n\nОжидаются столбцы:\n%v\n\nПолучены:\n%v\n\nОтправьте правильный файл с расписанием.",
					expectedHeaders, header),
			}
		}
	}

	return nil
}

func validateHeaders(actual, expected []string) bool {
	if len(actual) != len(expected) {
		return false
	}

	for i, exp := range expected {
		if actual[i] != exp {
			return false
		}
	}

	return true
}
