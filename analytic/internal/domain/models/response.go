package models

type TaskCountResponse struct {
	Count int `json:"count"`
}

type TotalTimeTasksResponse struct {
	Tasks []*Task `json:"tasks"`
}
