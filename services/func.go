package services

import (
	"encoding/json"
	"fmt"
	"myproject24/db"
	"myproject24/models"
	"time"
)

// ограничение времени приема
const (
	startHour = 8
	endHour   = 22
)

// создае временные метки для приема
func GenerateScheduleTimes(frequency int) []string {

	// Если лекарство принимают 1 раз в день — всегда в 08:00
	if frequency == 1 {
		return []string{"08:00"}
	}

	totalMinutes := (endHour - startHour) * 60
	interval := totalMinutes / (frequency - 1)

	var times []string
	for i := 0; i < frequency; i++ {
		minutes := startHour*60 + i*interval

		// Округляем вверх до ближайшего кратного 15 минут
		if minutes%15 != 0 {
			minutes = ((minutes / 15) + 1) * 15
		}

		hours := minutes / 60
		mins := minutes % 60
		timeStr := fmt.Sprintf("%02d:%02d", hours, mins)

		times = append(times, timeStr)
	}
	return times
}

func FilterTimesForToday(times []string, createdAt string, duration int, currentTime time.Time) ([]string, []string) {
	// Парсим дату создания расписания
	createdAtTime, _ := time.Parse("2006-01-02 15:04:05", createdAt)
	startDate := createdAtTime.Truncate(24 * time.Hour)
	endDate := startDate.Add(time.Duration(duration-1) * 24 * time.Hour)

	// Текущая дата
	currentDate := currentTime.Truncate(24 * time.Hour)

	var allTimes []string            // Все временные метки для текущего дня
	var remainingReceptions []string // Оставшиеся приёмы на текущий день

	for _, timeStr := range times {

		t, _ := time.Parse("15:04", timeStr)
		scheduledTime := time.Date(currentDate.Year(), currentDate.Month(), currentDate.Day(), t.Hour(), t.Minute(), 0, 0, time.Local)

		// Если текущий день выходит за пределы длительности курса, пропускаем время
		if currentDate.After(endDate) {
			continue
		}

		// Добавляем время в список всех временных меток для текущего дня
		allTimes = append(allTimes, timeStr)

		// Если время ещё не наступило, добавляем его в список оставшихся приёмов
		if scheduledTime.After(currentTime) {
			remainingReceptions = append(remainingReceptions, timeStr)
		}
	}

	return allTimes, remainingReceptions
}

func GetNextTakings(scheduleTimes []string) string {
	now := time.Now()
	currentHour := now.Hour()
	currentMinute := now.Minute()

	var nextTakings string
	// Преобразуем текущее время в минуты
	currentTime := currentHour*60 + currentMinute

	for _, timeStr := range scheduleTimes {
		var hour, min int
		fmt.Sscanf(timeStr, "%02d:%02d", &hour, &min)

		scheduledTime := hour*60 + min // Преобразуем время из расписания в минуты

		if scheduledTime > currentTime { // Проверяем, что время из расписания больше текущего
			nextTakings = timeStr
			break
		}
	}

	return nextTakings
}

func GetScheduleFromDB(userID, scheduleID string) (models.Schedule, error) {
	var schedule models.Schedule
	var timesJSON string

	err := db.DB.QueryRow("SELECT id, user_id, medicine, frequency, duration, created_at, times FROM schedules WHERE user_id = ? AND id = ?",
		userID, scheduleID).Scan(&schedule.ID, &schedule.UserID, &schedule.Medicine, &schedule.Frequency, &schedule.Duration, &schedule.CreatedAt, &timesJSON)

	if err != nil {
		return schedule, err
	}

	// Декодируем JSON-строку в массив строк
	if err := json.Unmarshal([]byte(timesJSON), &schedule.Times); err != nil {
		schedule.Times = GenerateScheduleTimes(schedule.Frequency)
	}

	return schedule, nil
}

// Проверка актуальности расписания
func IsScheduleActive(schedule models.Schedule) bool {
	createdAt, err := time.Parse("2006-01-02 15:04:05", schedule.CreatedAt)
	if err != nil {
		return false
	}

	duration := time.Duration(schedule.Duration) * 24 * time.Hour
	expirationDate := createdAt.Add(duration)
	today := time.Now()

	return !today.After(expirationDate)
}
