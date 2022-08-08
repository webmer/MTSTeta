package models

type UpdateTask struct {
	Name string `json:"name" example:"test task"`
	Text string `json:"text" example:"this is test task 1"`
}
