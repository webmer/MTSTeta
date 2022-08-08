package mocks

import (
	"context"

	"github.com/stretchr/testify/mock"
	"gitlab.com/g6834/team26/task/internal/domain/models"
	"gitlab.com/g6834/team26/task/internal/ports"
	"gitlab.com/g6834/team26/task/pkg/api"
)

type DbMock struct {
	mock.Mock
}

func (d *DbMock) List(ctx context.Context, login string) ([]*models.Task, error) {
	args := d.Called(ctx, login)
	return args.Get(0).([]*models.Task), args.Error(1)
}

func (d *DbMock) Run(ctx context.Context, t *models.Task) error {
	args := d.Called(ctx, t)
	return args.Error(0)
}

func (d *DbMock) Update(ctx context.Context, id, login, name, text string) error {
	args := d.Called(ctx, id, login, name, text)
	return args.Error(0)
}

func (d *DbMock) Delete(ctx context.Context, login, id string) error {
	args := d.Called(ctx, login, id)
	return args.Error(0)
}

func (d *DbMock) Approve(ctx context.Context, login, id, approvalLogin string) error {
	args := d.Called(ctx, login, id, approvalLogin)
	return args.Error(0)
}

func (d *DbMock) Decline(ctx context.Context, login, id, approvalLogin string) error {
	args := d.Called(ctx, login, id, approvalLogin)
	return args.Error(0)
}

func (d *DbMock) GetMessagesToSend(ctx context.Context) (map[int]models.KafkaAnalyticMessage, error) {
	args := d.Called(ctx)
	return args.Get(0).(map[int]models.KafkaAnalyticMessage), args.Error(1)
}

func (d *DbMock) GetEmailsToSend(ctx context.Context) ([]models.Email, error) {
	args := d.Called(ctx)
	return args.Get(0).([]models.Email), args.Error(1)
}

func (d *DbMock) UpdateMessageStatus(ctx context.Context, id int) error {
	args := d.Called(ctx, id)
	return args.Error(0)
}

func (d *DbMock) ChangeEmailStatusAndSendMessage(ctx context.Context, e models.Email, result bool) error {
	args := d.Called(ctx, e, result)
	return args.Error(0)
}

type GrpcAuthMock struct {
	mock.Mock
}

func (g *GrpcAuthMock) Validate(ctx context.Context, tokens ports.TokenPair) (*api.AuthResponse, error) {
	args := g.Called(ctx, tokens)
	return args.Get(0).(*api.AuthResponse), args.Error(1)
}

type AnalyticMessageSenderMock struct {
	mock.Mock
}

func (ams *AnalyticMessageSenderMock) ActionTask(ctx context.Context, m models.KafkaAnalyticMessage) error {
	args := ams.Called(ctx, m)
	return args.Error(0)
}

type EmailSenderMock struct {
	mock.Mock
}

func (es *EmailSenderMock) StartEmailWorkers(ctx context.Context) {
	es.Called(ctx)
	return
}

func (es *EmailSenderMock) SendEmail(e models.Email) error {
	args := es.Called(e)
	return args.Error(0)
}

func (es *EmailSenderMock) PushEmailToChan(e models.Email) {
	es.Called(e)
	return
}

func (es *EmailSenderMock) GetEmailResultChan() chan map[models.Email]bool {
	args := es.Called()
	return args.Get(0).(chan map[models.Email]bool)
}
