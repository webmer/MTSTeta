package models

import "database/sql"

type Approval struct {
	// Approved      bool   `json:"approved"`
	// Sent          bool   `json:"sent"`
	Approved      sql.NullBool `json:"approved"`
	Sent          sql.NullBool `json:"sent"`
	N             int          `json:"n" example:"2"`
	ApprovalLogin string       `json:"approvalLogin" example:"test626"`
}

// func (a *Approval) ChangeApprovedStatus(b bool) {
// 	a.Approved = b
// }

func (a *Approval) ChangeApprovedStatus(b bool) {
	a.Approved.Valid = true
	a.Approved.Bool = b
}

type Task struct {
	UUID           string      `json:"uuid" example:"eaca044f-5f02-4bc1-ba57-48845a473e42"`
	Name           string      `json:"name" example:"test task"`
	Text           string      `json:"text" example:"this is test task"`
	InitiatorLogin string      `json:"initiatorLogin" example:"test123"`
	Status         string      `json:"status" example:"created"`
	Approvals      []*Approval `json:"approvals"`
}
