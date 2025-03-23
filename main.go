package main

import (
	"log"
	"myproject24/config"
	"myproject24/db"
	"myproject24/handlers"

	"github.com/gin-gonic/gin"
)

func main() {
	// Загружаем конфигурацию
	config.LoadConfig()

	// Подключаем базу данных
	db.InitDB()
	defer db.CloseDB()

	// Устанавливить режим release на финальном этапе
	//gin.SetMode(gin.ReleaseMode)

	// Настраиваем роутер Gin
	r := gin.Default()
	handlers.SetupRoutes(r)

	// Запускаем сервер
	port := config.AppConfig.Port
	log.Printf("Сервер запущен на порту %s", port)
	r.Run(":" + port)

}
