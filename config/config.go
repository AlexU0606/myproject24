package config

// AppConfig хранит конфигурацию приложения
var AppConfig struct {
	Port string
}

// LoadConfig загружает переменные окружения
func LoadConfig() {
	AppConfig.Port = "8080"
}
