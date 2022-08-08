//go:build all || unit
// +build all unit

package models_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"gitlab.com/g6834/team26/task/internal/domain/models"
)

func TestRunTask(t *testing.T) {
	rt := models.RunTask{
		ApprovalLogins: []string{"test626", "zxcvb"},
		InitiatorLogin: "test123",
		Name:           "test task",
		Text:           "this is test task",
	}
	_, err := rt.CreateTask("test123")
	assert.ErrorIs(t, err, nil)
}
