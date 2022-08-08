package models

type Email struct {
	Id       int    `json:"id"`
	TaskUUID string `json:"task_uuid"`
	Reciever string `json:"reciever"`
	Type     string `json:"type"`
}
