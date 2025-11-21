package services

import (
	"encoding/csv"
	"fmt"
	"os"
	"strings"
)

type FileType string

const (
	FileTypeStudents FileType = "students"
	FileTypeTeachers FileType = "teachers"
	FileTypeSchedule FileType = "schedule"
)

const (
	errMsgInvalidCSV       = "Ошибка чтения CSV файла. Убедитесь что файл имеет правильный формат."
	errMsgEmptyFile        = "Файл пустой. Отправьте файл с данными."
	errMsgOnlyHeaders      = "Файл содержит только заголовки. Добавьте данные."
	errMsgInvalidStructure = "Неверная структура файла %s.\n\nОжидаются столбцы:\n%v\n\nПолучены:\n%v\n"
)

var expectedHeaders = map[FileType][]string{
	FileTypeStudents: {"User_id", "Last_name", "First_name", "Study_group"},
	FileTypeTeachers: {"User_id", "Last_name", "First_name"},
	FileTypeSchedule: {
		"subject_name", "type_name", "classroom", "group_name",
		"teacher_last_name", "teacher_first_name", "weekday", "start_time", "end_time",
	},
}

type ValidationError struct {
	Message string
}

func (e *ValidationError) Error() string {
	return e.Message
}

func newValidationError(msg string) *ValidationError {
	return &ValidationError{Message: msg}
}

func ValidateCSVStructure(filePath string, expectedType FileType) ([][]string, error) {
	records, err := readAndParseCSV(filePath)
	if err != nil {
		return nil, err
	}

	if err := validateRecordsNotEmpty(records); err != nil {
		return nil, err
	}

	if err := validateHeaders(records[0], expectedType); err != nil {
		return nil, err
	}

	return records, nil
}

func readAndParseCSV(filePath string) ([][]string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return nil, newValidationError(errMsgInvalidCSV)
	}

	return records, nil
}

func validateRecordsNotEmpty(records [][]string) error {
	if len(records) == 0 {
		return newValidationError(errMsgEmptyFile)
	}

	if len(records) == 1 {
		return newValidationError(errMsgOnlyHeaders)
	}

	return nil
}

func validateHeaders(actualHeaders []string, fileType FileType) error {
	expected, exists := expectedHeaders[fileType]
	if !exists {
		return fmt.Errorf("unknown file type: %s", fileType)
	}

	if !headersMatch(actualHeaders, expected) {
		return newValidationError(
			fmt.Sprintf(errMsgInvalidStructure, fileType, expected, actualHeaders),
		)
	}

	return nil
}

func headersMatch(actual, expected []string) bool {
	if len(actual) != len(expected) {
		return false
	}

	for i, exp := range expected {
		if len(actual) <= i || !strings.EqualFold(actual[i], exp) {
			return false
		}
	}

	return true
}
