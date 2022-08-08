package models

import (
	"database/sql"
	"time"
)

type Message struct {
	UUID        string
	UUIDMessage string
	Timestamp   time.Time
	Type        string
	Value       string
}

type TaskDB struct {
	UUID       string
	Login      string
	DateCreate time.Time
	DateAction time.Time
	Status     sql.NullBool
}

type MessageDB struct {
	UUID       string
	TaskUUID   string
	DateCreate time.Time
	Type       string
	Value      string
}

type Task struct {
	UUID      string `json:"uuid" example:"eaca044f-5f02-4bc1-ba57-48845a473e42"`
	TotalTime int    `json:"total_time" example:"2221"`
	Status    string `json:"status"  example:"false"`
}
