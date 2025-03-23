package models

// Schedule описывает модель расписания
type Schedule struct {
	ID        int      `json:"id"`
	UserID    string   `json:"user_id"`
	Medicine  string   `json:"medicine"`
	Frequency int      `json:"frequency"`
	Duration  int      `json:"duration"`
	CreatedAt string   `json:"created_at"`
	Times     []string `json:"times"`
}
