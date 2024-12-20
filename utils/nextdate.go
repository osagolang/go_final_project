package utils

import (
	"errors"
	"fmt"
	"go_final_project/constants"
	"strconv"
	"strings"
	"time"
)

// NextDate вычисляет следующую дату для задачи на основе правил повторения.
// Звёздочки (правила w и m) не реализовал (пока).
func NextDate(now time.Time, date string, repeat string) (string, error) {
	// Парсим начальную дату
	startDate, err := time.Parse(constants.DateFormat, date)
	if err != nil {
		return "", fmt.Errorf("invalid date format: %s", date)
	}

	// Если повторение пустое, возвращаем ошибку
	if repeat == "" {
		return "", errors.New("empty repeat rule")
	}

	// Разбираем правила
	ruleParts := strings.Split(repeat, " ")
	switch ruleParts[0] {
	case "d":
		// Правило "d <число>"
		if len(ruleParts) != 2 {
			return "", errors.New("invalid repeat format for 'd'")
		}
		days, err := strconv.Atoi(ruleParts[1])
		if err != nil || days <= 0 || days > 400 {
			return "", errors.New("invalid days in repeat rule")
		}

		// Вычисляем следующую дату
		for nextDate := startDate.AddDate(0, 0, days); ; nextDate = nextDate.AddDate(0, 0, days) {
			if nextDate.After(now) {
				return nextDate.Format(constants.DateFormat), nil
			}
		}

	case "y":
		// Правило "y"
		for nextDate := startDate.AddDate(1, 0, 0); ; nextDate = nextDate.AddDate(1, 0, 0) {
			if nextDate.After(now) {
				return nextDate.Format(constants.DateFormat), nil
			}
		}

	default:
		return "", errors.New("invalid or unsupported repeat rule")
	}
}

// NormalizeDate возвращает дату без времени (только год, месяц и день).
func NormalizeDate(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
}
