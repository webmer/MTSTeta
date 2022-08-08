package models

import (
	"database/sql"

	e "gitlab.com/g6834/team26/task/internal/domain/errors"
	"gitlab.com/g6834/team26/task/pkg/uuid"
)

type RunTask struct {
	ApprovalLogins []string `json:"approvalLogins" swaggertype:"array,string" example:"test626,zxcvb"`
	InitiatorLogin string   `json:"initiatorLogin" example:"test123"`
	Name           string   `json:"name" example:"test task"`
	Text           string   `json:"text" example:"this is test task 1"`
}

func (rt *RunTask) CreateTask(login string) (*Task, error) {
	if login != rt.InitiatorLogin {
		return &Task{}, e.ErrTokenLoginNotEqualInitiatorLogin
	}

	approvals := make([]*Approval, len(rt.ApprovalLogins))
	for idx, al := range rt.ApprovalLogins {
		approvals[idx] = &Approval{
			Approved:      sql.NullBool{Valid: false, Bool: false},
			Sent:          sql.NullBool{Valid: false, Bool: false},
			N:             idx + 1,
			ApprovalLogin: al,
		}
	}

	createdTask := &Task{
		UUID:           uuid.GenUUID(),
		InitiatorLogin: rt.InitiatorLogin,
		Name:           rt.Name,
		Text:           rt.Text,
		Status:         "created",
		Approvals:      approvals,
	}

	return createdTask, nil
}
