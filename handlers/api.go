package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"myproject24/db"
	"myproject24/models"
	"myproject24/services"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// SetupRoutes настраивает маршруты API
func SetupRoutes(r *gin.Engine) {
	r.POST("/schedule", createSchedule)
	r.GET("/schedules", getSchedules)
	r.GET("/schedule", getSchedule)
	r.GET("/next_takings", getNextTakings)
}

// Создание расписания
func createSchedule(c *gin.Context) {
	var schedule models.Schedule
	if err := c.ShouldBindJSON(&schedule); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	createdAt := time.Now()
	schedule.CreatedAt = createdAt.Format("2006-01-02 15:04:05")
	schedule.Times = services.GenerateScheduleTimes(schedule.Frequency)

	log.Printf("Calculated times: %v", schedule.Times)

	timesJSON, _ := json.Marshal(schedule.Times)
	log.Printf("Generated JSON for times: %s", string(timesJSON))

	result, err := db.DB.Exec("INSERT INTO schedules (user_id, medicine, frequency, duration, created_at, times) VALUES (?, ?, ?, ?, ?, ?)",
		schedule.UserID, schedule.Medicine, schedule.Frequency, schedule.Duration, schedule.CreatedAt, string(timesJSON))

	if err != nil {
		log.Printf("Database insert error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error", "details": err.Error()})
		return
	}

	id, _ := result.LastInsertId()
	c.JSON(http.StatusOK, gin.H{"schedule_id": id})
}

// Получение всех расписаний пользователя
func getSchedules(c *gin.Context) {
	userID := c.Query("user_id")
	rows, err := db.DB.Query("SELECT id FROM schedules WHERE user_id = ?", userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}
	defer rows.Close()

	var schedules []int
	for rows.Next() {
		var id int
		rows.Scan(&id)
		schedules = append(schedules, id)
	}

	c.JSON(http.StatusOK, gin.H{"schedules": schedules})
}

// Получение конкретного расписания с проверкой на его актуальность
func getSchedule(c *gin.Context) {
	userID, scheduleID := c.Query("user_id"), c.Query("schedule_id")

	// Получаем расписание из базы данных
	schedule, err := services.GetScheduleFromDB(userID, scheduleID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Schedule not found", "details": err.Error()})
		return
	}

	// Проверяем актуальность расписания
	if !services.IsScheduleActive(schedule) {
		c.JSON(http.StatusOK, gin.H{"message": "Расписание не актуально"})
		return
	}

	// Фильтруем временные метки для текущего дня
	currentTime := time.Now()
	allTimes, remainingMedications := services.FilterTimesForToday(schedule.Times, schedule.CreatedAt, schedule.Duration, currentTime)
	schedule.Times = allTimes

	// Формируем ответ

	c.JSON(http.StatusOK, gin.H{
		"schedules":                       schedule,
		"remaining medications for today": remainingMedications,
	})

}

// Получение ближайших приемов
func getNextTakings(c *gin.Context) {
	userID, scheduleID := c.Query("user_id"), c.Query("schedule_id")

	schedule, err := services.GetScheduleFromDB(userID, scheduleID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Schedule not found"})
		return
	}

	// Получаем следующее время приема
	nextTaking := services.GetNextTakings(schedule.Times)

	// Если на сегодня приемов больше нет
	if nextTaking == "" {
		nextTaking = fmt.Sprintf("на сегодня, приемов лекарства %s больше нет", schedule.Medicine)
	}

	c.JSON(http.StatusOK, gin.H{
		"medicine":   schedule.Medicine,
		"nextTaking": nextTaking,
	})
}
