//go:build all || unit
// +build all unit

package http_test

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	h "gitlab.com/g6834/team26/task/internal/adapters/http"
	e "gitlab.com/g6834/team26/task/internal/domain/errors"
	"gitlab.com/g6834/team26/task/internal/domain/models"
	"gitlab.com/g6834/team26/task/internal/domain/task"
	"gitlab.com/g6834/team26/task/internal/ports"
	"gitlab.com/g6834/team26/task/pkg/api"
	"gitlab.com/g6834/team26/task/pkg/config"
	"gitlab.com/g6834/team26/task/pkg/logger"
	"gitlab.com/g6834/team26/task/pkg/mocks"
)

type authTestSuite struct {
	suite.Suite

	srv            *h.Server
	db             *mocks.DbMock
	gAuth          *mocks.GrpcAuthMock
	analyticSender *mocks.AnalyticMessageSenderMock
	emailSender    *mocks.EmailSenderMock
}

func TestAuthTestSuite(t *testing.T) {
	suite.Run(t, &authTestSuite{})
}

func (s *authTestSuite) SetupTest() {
	// l := zerolog.Nop()
	l := logger.New()

	c, err := config.New()
	if err != nil {
		s.Suite.T().Errorf("Error parsing env: %s", err)
	}

	s.db = new(mocks.DbMock)
	s.gAuth = new(mocks.GrpcAuthMock)
	s.analyticSender = new(mocks.AnalyticMessageSenderMock)
	s.emailSender = new(mocks.EmailSenderMock)

	taskS := task.New(s.db, s.gAuth, s.analyticSender, s.emailSender)

	s.srv, err = h.New(l, taskS, c)
	if err != nil {
		s.Suite.T().Errorf("db init failed: %s", err)
		s.Suite.T().FailNow()
	}

	go s.srv.Start(context.Background())
}

func (s *authTestSuite) TearDownTest() {
	_ = s.srv.Stop(context.Background())
}

func (s *authTestSuite) TestListHandlerSuccess() {
	// ctx := context.Background()
	s.emailSender.On("StartEmailWorkers", mock.Anything).Return()
	s.emailSender.On("GetEmailResultChan").Return(make(chan map[models.Email]bool))
	s.db.On("List", mock.Anything, "test123").Return([]*models.Task{&models.Task{UUID: "66f5b904-3f54-4da4-ba74-6dfdf8d72efb",
		Name:           "test",
		Text:           "this is test task",
		InitiatorLogin: "test123",
		Status:         "created",
		Approvals: []*models.Approval{&models.Approval{ApprovalLogin: "test626",
			N:        2,
			Sent:     sql.NullBool{Valid: true, Bool: false},
			Approved: sql.NullBool{Valid: false, Bool: false}}}}}, nil)
	s.db.On("GetMessagesToSend", mock.Anything).Return(map[int]models.KafkaAnalyticMessage{}, nil)
	s.db.On("UpdateMessageStatus", mock.Anything, mock.Anything).Return(nil)
	s.gAuth.On("Validate", mock.Anything, ports.TokenPair{
		AccessToken: ports.TokenPairVal{
			Value: "access_token",
			// Expires: time.Now().Add(time.Hour),
		},
		RefreshToken: ports.TokenPairVal{
			Value: "refresh_token",
			// Expires: time.Now().Add(time.Hour),
		}}).Return(&api.AuthResponse{Result: true, Login: "test123", AccessToken: new(api.Token), RefreshToken: new(api.Token)}, nil)

	bodyReq := strings.NewReader("")

	req, err := http.NewRequest("GET", "http://localhost:3000/task/v1/tasks/", bodyReq)
	s.NoError(err)
	req.AddCookie(&http.Cookie{Name: "access_token", Value: "access_token"})
	req.AddCookie(&http.Cookie{Name: "refresh_token", Value: "refresh_token"})

	client := http.Client{}
	response, err := client.Do(req)

	// log.Println(err)
	// data, err := ioutil.ReadAll(response.Body)
	// log.Println(string(data))
	s.NoError(err)
	s.Equal(http.StatusOK, response.StatusCode)
	response.Body.Close()
}

func (s *authTestSuite) TestListHandlerForbidden() {
	// ctx := context.Background()
	s.emailSender.On("StartEmailWorkers", mock.Anything).Return()
	s.emailSender.On("GetEmailResultChan").Return(make(chan map[models.Email]bool))
	s.db.On("List", mock.Anything, "test123").Return([]*models.Task{&models.Task{UUID: "66f5b904-3f54-4da4-ba74-6dfdf8d72efb",
		Name:           "test",
		Text:           "this is test task",
		InitiatorLogin: "test123",
		Status:         "created",
		Approvals: []*models.Approval{&models.Approval{ApprovalLogin: "test626",
			N:        2,
			Sent:     sql.NullBool{Valid: true, Bool: false},
			Approved: sql.NullBool{Valid: false, Bool: false}}}}}, nil)
	s.db.On("GetMessagesToSend", mock.Anything).Return(map[int]models.KafkaAnalyticMessage{}, nil)
	s.db.On("UpdateMessageStatus", mock.Anything, mock.Anything).Return(nil)
	s.gAuth.On("Validate", mock.Anything, ports.TokenPair{
		AccessToken: ports.TokenPairVal{
			Value: "access_token",
			// Expires: time.Now().Add(time.Hour),
		},
		RefreshToken: ports.TokenPairVal{
			Value: "refresh_token",
			// Expires: time.Now().Add(time.Hour),
		}}).Return(&api.AuthResponse{Result: false, Login: "test123", AccessToken: new(api.Token), RefreshToken: new(api.Token)}, e.ErrAuthFailed)

	bodyReq := strings.NewReader("")

	req, err := http.NewRequest("GET", "http://localhost:3000/task/v1/tasks/", bodyReq)
	s.NoError(err)
	req.AddCookie(&http.Cookie{Name: "access_token", Value: "access_token"})
	req.AddCookie(&http.Cookie{Name: "refresh_token", Value: "refresh_token"})

	client := http.Client{}
	response, err := client.Do(req)

	// log.Println(err)
	// data, err := ioutil.ReadAll(response.Body)
	// log.Println(string(data))
	s.NoError(err)
	s.Equal(http.StatusForbidden, response.StatusCode)
	response.Body.Close()
}

func (s *authTestSuite) TestRunHandlerSuccess() {
	// ctx := context.Background()
	s.emailSender.On("StartEmailWorkers", mock.Anything).Return()
	s.emailSender.On("GetEmailResultChan").Return(make(chan map[models.Email]bool))
	s.db.On("Run", mock.Anything, mock.Anything).Return(nil)
	s.db.On("GetMessagesToSend", mock.Anything).Return(map[int]models.KafkaAnalyticMessage{}, nil)
	s.db.On("UpdateMessageStatus", mock.Anything, mock.Anything).Return(nil)
	s.gAuth.On("Validate", mock.Anything, ports.TokenPair{
		AccessToken: ports.TokenPairVal{
			Value: "access_token",
			// Expires: time.Now().Add(time.Hour),
		},
		RefreshToken: ports.TokenPairVal{
			Value: "refresh_token",
			// Expires: time.Now().Add(time.Hour),
		}}).Return(&api.AuthResponse{Result: true, Login: "test123", AccessToken: new(api.Token), RefreshToken: new(api.Token)}, nil)
	s.analyticSender.On("ActionTask", mock.Anything, mock.Anything).Return(nil)

	bodyReq := strings.NewReader("{\"approvalLogins\": [\"test626\",\"zxcvb\"],\"initiatorLogin\": \"test123\"}")

	req, err := http.NewRequest("POST", "http://localhost:3000/task/v1/tasks/run", bodyReq)
	s.NoError(err)
	req.AddCookie(&http.Cookie{Name: "access_token", Value: "access_token"})
	req.AddCookie(&http.Cookie{Name: "refresh_token", Value: "refresh_token"})

	client := http.Client{}
	response, err := client.Do(req)

	// log.Println(err)
	// data, err := ioutil.ReadAll(response.Body)
	// log.Println(string(data))
	s.NoError(err)
	s.Equal(http.StatusOK, response.StatusCode)
	response.Body.Close()
}

func (s *authTestSuite) TestRunHandlerBadRequest() {
	// ctx := context.Background()
	s.emailSender.On("StartEmailWorkers", mock.Anything).Return()
	s.emailSender.On("GetEmailResultChan").Return(make(chan map[models.Email]bool))
	s.db.On("Run", mock.Anything, mock.Anything).Return(e.ErrInvalidJsonBody)
	s.db.On("GetMessagesToSend", mock.Anything).Return(map[int]models.KafkaAnalyticMessage{}, nil)
	s.db.On("UpdateMessageStatus", mock.Anything, mock.Anything).Return(nil)
	s.gAuth.On("Validate", mock.Anything, ports.TokenPair{
		AccessToken: ports.TokenPairVal{
			Value: "access_token",
			// Expires: time.Now().Add(time.Hour),
		},
		RefreshToken: ports.TokenPairVal{
			Value: "refresh_token",
			// Expires: time.Now().Add(time.Hour),
		}}).Return(&api.AuthResponse{Result: true, Login: "test123", AccessToken: new(api.Token), RefreshToken: new(api.Token)}, nil)
	s.analyticSender.On("ActionTask", mock.Anything, mock.Anything).Return(nil)

	bodyReq := strings.NewReader("{\"approvalLogins\": {\"test626\": \"\", \"zxcvb\": \"\"},\"initiatorLogin\": \"test123\"}")

	req, err := http.NewRequest("POST", "http://localhost:3000/task/v1/tasks/run", bodyReq)
	s.NoError(err)
	req.AddCookie(&http.Cookie{Name: "access_token", Value: "access_token"})
	req.AddCookie(&http.Cookie{Name: "refresh_token", Value: "refresh_token"})

	client := http.Client{}
	response, err := client.Do(req)

	// log.Println(err)
	// data, err := ioutil.ReadAll(response.Body)
	// log.Println(string(data))
	s.NoError(err)
	s.Equal(http.StatusBadRequest, response.StatusCode)
	response.Body.Close()
}

func (s *authTestSuite) TestRunHandlerForbidden() {
	// ctx := context.Background()
	s.emailSender.On("StartEmailWorkers", mock.Anything).Return()
	s.emailSender.On("GetEmailResultChan").Return(make(chan map[models.Email]bool))
	s.db.On("Run", mock.Anything, mock.Anything).Return(nil)
	s.db.On("GetMessagesToSend", mock.Anything).Return(map[int]models.KafkaAnalyticMessage{}, nil)
	s.db.On("UpdateMessageStatus", mock.Anything, mock.Anything).Return(nil)
	s.gAuth.On("Validate", mock.Anything, ports.TokenPair{
		AccessToken: ports.TokenPairVal{
			Value: "access_token",
			// Expires: time.Now().Add(time.Hour),
		},
		RefreshToken: ports.TokenPairVal{
			Value: "refresh_token",
			// Expires: time.Now().Add(time.Hour),
		}}).Return(&api.AuthResponse{Result: false, Login: "test123", AccessToken: new(api.Token), RefreshToken: new(api.Token)}, e.ErrAuthFailed)
	s.analyticSender.On("ActionTask", mock.Anything, mock.Anything).Return(nil)

	bodyReq := strings.NewReader("{\"approvalLogins\": [\"test626\",\"zxcvb\"],\"initiatorLogin\": \"test123\"}")

	req, err := http.NewRequest("POST", "http://localhost:3000/task/v1/tasks/run", bodyReq)
	s.NoError(err)
	req.AddCookie(&http.Cookie{Name: "access_token", Value: "access_token"})
	req.AddCookie(&http.Cookie{Name: "refresh_token", Value: "refresh_token"})

	client := http.Client{}
	response, err := client.Do(req)

	// log.Println(err)
	// data, err := ioutil.ReadAll(response.Body)
	// log.Println(string(data))
	s.NoError(err)
	s.Equal(http.StatusForbidden, response.StatusCode)
	response.Body.Close()
}

func (s *authTestSuite) TestUpdateHandlerSuccess() {
	// ctx := context.Background()
	s.emailSender.On("StartEmailWorkers", mock.Anything).Return()
	s.emailSender.On("GetEmailResultChan").Return(make(chan map[models.Email]bool))
	s.db.On("Update", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
	s.db.On("GetMessagesToSend", mock.Anything).Return(map[int]models.KafkaAnalyticMessage{}, nil)
	s.db.On("UpdateMessageStatus", mock.Anything, mock.Anything).Return(nil)
	s.gAuth.On("Validate", mock.Anything, ports.TokenPair{
		AccessToken: ports.TokenPairVal{
			Value: "access_token",
			// Expires: time.Now().Add(time.Hour),
		},
		RefreshToken: ports.TokenPairVal{
			Value: "refresh_token",
			// Expires: time.Now().Add(time.Hour),
		}}).Return(&api.AuthResponse{Result: true, Login: "test123", AccessToken: new(api.Token), RefreshToken: new(api.Token)}, nil)
	s.analyticSender.On("ActionTask", mock.Anything, mock.Anything).Return(nil)

	bodyReq := strings.NewReader("{\"name\": \"name update\", \"text\": \"text update\"}")
	uuid := "66f5b904-3f54-4da4-ba74-6dfdf8d72efb"

	req, err := http.NewRequest("PUT", fmt.Sprintf("http://localhost:3000/task/v1/tasks/%s", uuid), bodyReq)
	s.NoError(err)
	req.AddCookie(&http.Cookie{Name: "access_token", Value: "access_token"})
	req.AddCookie(&http.Cookie{Name: "refresh_token", Value: "refresh_token"})

	client := http.Client{}
	response, err := client.Do(req)

	// log.Println(err)
	// data, err := ioutil.ReadAll(response.Body)
	// log.Println(string(data))
	s.NoError(err)
	s.Equal(http.StatusOK, response.StatusCode)
	response.Body.Close()
}

func (s *authTestSuite) TestUpdateHandlerBadRequest() {
	// ctx := context.Background()
	s.emailSender.On("StartEmailWorkers", mock.Anything).Return()
	s.emailSender.On("GetEmailResultChan").Return(make(chan map[models.Email]bool))
	s.db.On("Update", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(e.ErrInvalidJsonBody)
	s.db.On("GetMessagesToSend", mock.Anything).Return(map[int]models.KafkaAnalyticMessage{}, nil)
	s.db.On("UpdateMessageStatus", mock.Anything, mock.Anything).Return(nil)
	s.gAuth.On("Validate", mock.Anything, ports.TokenPair{
		AccessToken: ports.TokenPairVal{
			Value: "access_token",
			// Expires: time.Now().Add(time.Hour),
		},
		RefreshToken: ports.TokenPairVal{
			Value: "refresh_token",
			// Expires: time.Now().Add(time.Hour),
		}}).Return(&api.AuthResponse{Result: true, Login: "test123", AccessToken: new(api.Token), RefreshToken: new(api.Token)}, nil)
	s.analyticSender.On("ActionTask", mock.Anything, mock.Anything).Return(nil)

	bodyReq := strings.NewReader("{\"name\": \"name update\", \"text\": \"text update\"}")
	uuid := "66f5b904-3f54-4da4-ba74-6dfdf8d72efb"

	req, err := http.NewRequest("PUT", fmt.Sprintf("http://localhost:3000/task/v1/tasks/%s", uuid), bodyReq)
	s.NoError(err)
	req.AddCookie(&http.Cookie{Name: "access_token", Value: "access_token"})
	req.AddCookie(&http.Cookie{Name: "refresh_token", Value: "refresh_token"})

	client := http.Client{}
	response, err := client.Do(req)

	// log.Println(err)
	// data, err := ioutil.ReadAll(response.Body)
	// log.Println(string(data))
	s.NoError(err)
	s.Equal(http.StatusBadRequest, response.StatusCode)
	response.Body.Close()
}

func (s *authTestSuite) TestUpdateHandlerForbidden() {
	// ctx := context.Background()
	s.emailSender.On("StartEmailWorkers", mock.Anything).Return()
	s.emailSender.On("GetEmailResultChan").Return(make(chan map[models.Email]bool))
	s.db.On("Update", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
	s.db.On("GetMessagesToSend", mock.Anything).Return(map[int]models.KafkaAnalyticMessage{}, nil)
	s.db.On("UpdateMessageStatus", mock.Anything, mock.Anything).Return(nil)
	s.gAuth.On("Validate", mock.Anything, ports.TokenPair{
		AccessToken: ports.TokenPairVal{
			Value: "access_token",
			// Expires: time.Now().Add(time.Hour),
		},
		RefreshToken: ports.TokenPairVal{
			Value: "refresh_token",
			// Expires: time.Now().Add(time.Hour),
		}}).Return(&api.AuthResponse{Result: false, Login: "test123", AccessToken: new(api.Token), RefreshToken: new(api.Token)}, e.ErrAuthFailed)
	s.analyticSender.On("ActionTask", mock.Anything, mock.Anything).Return(nil)

	bodyReq := strings.NewReader("{\"name\": \"name update\", \"text\": \"text update\"}")
	uuid := "66f5b904-3f54-4da4-ba74-6dfdf8d72efb"

	req, err := http.NewRequest("PUT", fmt.Sprintf("http://localhost:3000/task/v1/tasks/%s", uuid), bodyReq)
	s.NoError(err)
	req.AddCookie(&http.Cookie{Name: "access_token", Value: "access_token"})
	req.AddCookie(&http.Cookie{Name: "refresh_token", Value: "refresh_token"})

	client := http.Client{}
	response, err := client.Do(req)

	// log.Println(err)
	// data, err := ioutil.ReadAll(response.Body)
	// log.Println(string(data))
	s.NoError(err)
	s.Equal(http.StatusForbidden, response.StatusCode)
	response.Body.Close()
}

func (s *authTestSuite) TestApproveHandlerSuccess() {
	// ctx := context.Background()
	s.emailSender.On("StartEmailWorkers", mock.Anything).Return()
	s.emailSender.On("GetEmailResultChan").Return(make(chan map[models.Email]bool))
	s.db.On("Approve", mock.Anything, "test123", "66f5b904-3f54-4da4-ba74-6dfdf8d72efb", "test626").Return(nil)
	s.db.On("GetMessagesToSend", mock.Anything).Return(map[int]models.KafkaAnalyticMessage{}, nil)
	s.db.On("UpdateMessageStatus", mock.Anything, mock.Anything).Return(nil)
	s.gAuth.On("Validate", mock.Anything, ports.TokenPair{
		AccessToken: ports.TokenPairVal{
			Value: "access_token",
			// Expires: time.Now().Add(time.Hour),
		},
		RefreshToken: ports.TokenPairVal{
			Value: "refresh_token",
			// Expires: time.Now().Add(time.Hour),
		}}).Return(&api.AuthResponse{Result: true, Login: "test123", AccessToken: new(api.Token), RefreshToken: new(api.Token)}, nil)
	s.analyticSender.On("ActionTask", mock.Anything, mock.Anything).Return(nil)

	approvalLogin := "test626"
	uuid := "66f5b904-3f54-4da4-ba74-6dfdf8d72efb"
	bodyReq := strings.NewReader("")

	req, err := http.NewRequest("POST", fmt.Sprintf("http://localhost:3000/task/v1/tasks/%s/approve/%s", uuid, approvalLogin), bodyReq)
	s.NoError(err)
	req.AddCookie(&http.Cookie{Name: "access_token", Value: "access_token"})
	req.AddCookie(&http.Cookie{Name: "refresh_token", Value: "refresh_token"})

	client := http.Client{}
	response, err := client.Do(req)

	// log.Println(err)
	// data, err := ioutil.ReadAll(response.Body)
	// log.Println(string(data))
	s.NoError(err)
	s.Equal(http.StatusOK, response.StatusCode)
	response.Body.Close()
}

func (s *authTestSuite) TestApproveHandlerForbidden() {
	// ctx := context.Background()
	s.emailSender.On("StartEmailWorkers", mock.Anything).Return()
	s.emailSender.On("GetEmailResultChan").Return(make(chan map[models.Email]bool))
	s.db.On("Approve", mock.Anything, "test123", "66f5b904-3f54-4da4-ba74-6dfdf8d72efb", "test626").Return(nil)
	s.db.On("GetMessagesToSend", mock.Anything).Return(map[int]models.KafkaAnalyticMessage{}, nil)
	s.db.On("UpdateMessageStatus", mock.Anything, mock.Anything).Return(nil)
	s.gAuth.On("Validate", mock.Anything, ports.TokenPair{
		AccessToken: ports.TokenPairVal{
			Value: "access_token",
			// Expires: time.Now().Add(time.Hour),
		},
		RefreshToken: ports.TokenPairVal{
			Value: "refresh_token",
			// Expires: time.Now().Add(time.Hour),
		}}).Return(&api.AuthResponse{Result: false, Login: "test123", AccessToken: new(api.Token), RefreshToken: new(api.Token)}, e.ErrAuthFailed)
	s.analyticSender.On("ActionTask", mock.Anything, mock.Anything).Return(nil)

	approvalLogin := "test626"
	uuid := "66f5b904-3f54-4da4-ba74-6dfdf8d72efb"
	bodyReq := strings.NewReader("")

	req, err := http.NewRequest("POST", fmt.Sprintf("http://localhost:3000/task/v1/tasks/%s/approve/%s", uuid, approvalLogin), bodyReq)
	s.NoError(err)
	req.AddCookie(&http.Cookie{Name: "access_token", Value: "access_token"})
	req.AddCookie(&http.Cookie{Name: "refresh_token", Value: "refresh_token"})

	client := http.Client{}
	response, err := client.Do(req)

	// log.Println(err)
	// data, err := ioutil.ReadAll(response.Body)
	// log.Println(string(data))
	s.NoError(err)
	s.Equal(http.StatusForbidden, response.StatusCode)
	response.Body.Close()
}

func (s *authTestSuite) TestApproveHandlerNotFound() {
	// ctx := context.Background()
	s.emailSender.On("StartEmailWorkers", mock.Anything).Return()
	s.emailSender.On("GetEmailResultChan").Return(make(chan map[models.Email]bool))
	s.db.On("Approve", mock.Anything, "test123", "66f5b904-3f54-4da4-ba74-6dfdf8d72efb", "test626").Return(e.ErrNotFound)
	s.db.On("GetMessagesToSend", mock.Anything).Return(map[int]models.KafkaAnalyticMessage{}, nil)
	s.db.On("UpdateMessageStatus", mock.Anything, mock.Anything).Return(nil)
	s.gAuth.On("Validate", mock.Anything, ports.TokenPair{
		AccessToken: ports.TokenPairVal{
			Value: "access_token",
			// Expires: time.Now().Add(time.Hour),
		},
		RefreshToken: ports.TokenPairVal{
			Value: "refresh_token",
			// Expires: time.Now().Add(time.Hour),
		}}).Return(&api.AuthResponse{Result: true, Login: "test123", AccessToken: new(api.Token), RefreshToken: new(api.Token)}, nil)
	s.analyticSender.On("ActionTask", mock.Anything, mock.Anything).Return(nil)

	approvalLogin := "test626"
	uuid := "66f5b904-3f54-4da4-ba74-6dfdf8d72efb"
	bodyReq := strings.NewReader("")

	req, err := http.NewRequest("POST", fmt.Sprintf("http://localhost:3000/task/v1/tasks/%s/approve/%s", uuid, approvalLogin), bodyReq)
	s.NoError(err)
	req.AddCookie(&http.Cookie{Name: "access_token", Value: "access_token"})
	req.AddCookie(&http.Cookie{Name: "refresh_token", Value: "refresh_token"})

	client := http.Client{}
	response, err := client.Do(req)

	// log.Println(err)
	// data, err := ioutil.ReadAll(response.Body)
	// log.Println(string(data))
	s.NoError(err)
	s.Equal(http.StatusNotFound, response.StatusCode)
	response.Body.Close()
}

func (s *authTestSuite) TestDeclineHandlerSuccess() {
	// ctx := context.Background()
	s.emailSender.On("StartEmailWorkers", mock.Anything).Return()
	s.emailSender.On("GetEmailResultChan").Return(make(chan map[models.Email]bool))
	s.db.On("Decline", mock.Anything, "test123", "66f5b904-3f54-4da4-ba74-6dfdf8d72efb", "test626").Return(nil)
	s.db.On("GetMessagesToSend", mock.Anything).Return(map[int]models.KafkaAnalyticMessage{}, nil)
	s.db.On("UpdateMessageStatus", mock.Anything, mock.Anything).Return(nil)
	s.gAuth.On("Validate", mock.Anything, ports.TokenPair{
		AccessToken: ports.TokenPairVal{
			Value: "access_token",
			// Expires: time.Now().Add(time.Hour),
		},
		RefreshToken: ports.TokenPairVal{
			Value: "refresh_token",
			// Expires: time.Now().Add(time.Hour),
		}}).Return(&api.AuthResponse{Result: true, Login: "test123", AccessToken: new(api.Token), RefreshToken: new(api.Token)}, nil)
	s.analyticSender.On("ActionTask", mock.Anything, mock.Anything).Return(nil)

	approvalLogin := "test626"
	uuid := "66f5b904-3f54-4da4-ba74-6dfdf8d72efb"
	bodyReq := strings.NewReader("")

	req, err := http.NewRequest("POST", fmt.Sprintf("http://localhost:3000/task/v1/tasks/%s/decline/%s", uuid, approvalLogin), bodyReq)
	s.NoError(err)
	req.AddCookie(&http.Cookie{Name: "access_token", Value: "access_token"})
	req.AddCookie(&http.Cookie{Name: "refresh_token", Value: "refresh_token"})

	client := http.Client{}
	response, err := client.Do(req)

	// log.Println(err)
	// data, err := ioutil.ReadAll(response.Body)
	// log.Println(string(data))
	s.NoError(err)
	s.Equal(http.StatusOK, response.StatusCode)
	response.Body.Close()
}

func (s *authTestSuite) TestDeclineHandlerForbidden() {
	// ctx := context.Background()
	s.emailSender.On("StartEmailWorkers", mock.Anything).Return()
	s.emailSender.On("GetEmailResultChan").Return(make(chan map[models.Email]bool))
	s.db.On("Decline", mock.Anything, "test123", "66f5b904-3f54-4da4-ba74-6dfdf8d72efb", "test626").Return(nil)
	s.db.On("GetMessagesToSend", mock.Anything).Return(map[int]models.KafkaAnalyticMessage{}, nil)
	s.db.On("UpdateMessageStatus", mock.Anything, mock.Anything).Return(nil)
	s.gAuth.On("Validate", mock.Anything, ports.TokenPair{
		AccessToken: ports.TokenPairVal{
			Value: "access_token",
			// Expires: time.Now().Add(time.Hour),
		},
		RefreshToken: ports.TokenPairVal{
			Value: "refresh_token",
			// Expires: time.Now().Add(time.Hour),
		}}).Return(&api.AuthResponse{Result: false, Login: "test123", AccessToken: new(api.Token), RefreshToken: new(api.Token)}, e.ErrAuthFailed)
	s.analyticSender.On("ActionTask", mock.Anything, mock.Anything).Return(nil)

	approvalLogin := "test626"
	uuid := "66f5b904-3f54-4da4-ba74-6dfdf8d72efb"
	bodyReq := strings.NewReader("")

	req, err := http.NewRequest("POST", fmt.Sprintf("http://localhost:3000/task/v1/tasks/%s/decline/%s", uuid, approvalLogin), bodyReq)
	s.NoError(err)
	req.AddCookie(&http.Cookie{Name: "access_token", Value: "access_token"})
	req.AddCookie(&http.Cookie{Name: "refresh_token", Value: "refresh_token"})

	client := http.Client{}
	response, err := client.Do(req)

	// log.Println(err)
	// data, err := ioutil.ReadAll(response.Body)
	// log.Println(string(data))
	s.NoError(err)
	s.Equal(http.StatusForbidden, response.StatusCode)
	response.Body.Close()
}

func (s *authTestSuite) TestDeclineHandlerNotFound() {
	// ctx := context.Background()
	s.emailSender.On("StartEmailWorkers", mock.Anything).Return()
	s.emailSender.On("GetEmailResultChan").Return(make(chan map[models.Email]bool))
	s.db.On("Decline", mock.Anything, "test123", "66f5b904-3f54-4da4-ba74-6dfdf8d72efb", "test626").Return(e.ErrNotFound)
	s.db.On("GetMessagesToSend", mock.Anything).Return(map[int]models.KafkaAnalyticMessage{}, nil)
	s.db.On("UpdateMessageStatus", mock.Anything, mock.Anything).Return(nil)
	s.gAuth.On("Validate", mock.Anything, ports.TokenPair{
		AccessToken: ports.TokenPairVal{
			Value: "access_token",
			// Expires: time.Now().Add(time.Hour),
		},
		RefreshToken: ports.TokenPairVal{
			Value: "refresh_token",
			// Expires: time.Now().Add(time.Hour),
		}}).Return(&api.AuthResponse{Result: true, Login: "test123", AccessToken: new(api.Token), RefreshToken: new(api.Token)}, nil)
	s.analyticSender.On("ActionTask", mock.Anything, mock.Anything).Return(nil)

	approvalLogin := "test626"
	uuid := "66f5b904-3f54-4da4-ba74-6dfdf8d72efb"
	bodyReq := strings.NewReader("")

	req, err := http.NewRequest("POST", fmt.Sprintf("http://localhost:3000/task/v1/tasks/%s/decline/%s", uuid, approvalLogin), bodyReq)
	s.NoError(err)
	req.AddCookie(&http.Cookie{Name: "access_token", Value: "access_token"})
	req.AddCookie(&http.Cookie{Name: "refresh_token", Value: "refresh_token"})

	client := http.Client{}
	response, err := client.Do(req)

	// log.Println(err)
	// data, err := ioutil.ReadAll(response.Body)
	// log.Println(string(data))
	s.NoError(err)
	s.Equal(http.StatusNotFound, response.StatusCode)
	response.Body.Close()
}

func (s *authTestSuite) TestDeleteHandlerSuccess() {
	// ctx := context.Background()
	s.emailSender.On("StartEmailWorkers", mock.Anything).Return()
	s.emailSender.On("GetEmailResultChan").Return(make(chan map[models.Email]bool))
	s.db.On("Delete", mock.Anything, "test123", "66f5b904-3f54-4da4-ba74-6dfdf8d72efb").Return(nil)
	s.db.On("GetMessagesToSend", mock.Anything).Return(map[int]models.KafkaAnalyticMessage{}, nil)
	s.db.On("UpdateMessageStatus", mock.Anything, mock.Anything).Return(nil)
	s.gAuth.On("Validate", mock.Anything, ports.TokenPair{
		AccessToken: ports.TokenPairVal{
			Value: "access_token",
			// Expires: time.Now().Add(time.Hour),
		},
		RefreshToken: ports.TokenPairVal{
			Value: "refresh_token",
			// Expires: time.Now().Add(time.Hour),
		}}).Return(&api.AuthResponse{Result: true, Login: "test123", AccessToken: new(api.Token), RefreshToken: new(api.Token)}, nil)
	s.analyticSender.On("ActionTask", mock.Anything, mock.Anything).Return(nil)

	uuid := "66f5b904-3f54-4da4-ba74-6dfdf8d72efb"
	bodyReq := strings.NewReader("")

	req, err := http.NewRequest("DELETE", fmt.Sprintf("http://localhost:3000/task/v1/tasks/%s", uuid), bodyReq)
	s.NoError(err)
	req.AddCookie(&http.Cookie{Name: "access_token", Value: "access_token"})
	req.AddCookie(&http.Cookie{Name: "refresh_token", Value: "refresh_token"})

	client := http.Client{}
	response, err := client.Do(req)

	// log.Println(err)
	// data, err := ioutil.ReadAll(response.Body)
	// log.Println(string(data))
	s.NoError(err)
	s.Equal(http.StatusOK, response.StatusCode)
	response.Body.Close()
}

func (s *authTestSuite) TestDeleteHandlerForbidden() {
	// ctx := context.Background()
	s.emailSender.On("StartEmailWorkers", mock.Anything).Return()
	s.emailSender.On("GetEmailResultChan").Return(make(chan map[models.Email]bool))
	s.db.On("Delete", mock.Anything, "test123", "66f5b904-3f54-4da4-ba74-6dfdf8d72efb").Return(nil)
	s.db.On("GetMessagesToSend", mock.Anything).Return(map[int]models.KafkaAnalyticMessage{}, nil)
	s.db.On("UpdateMessageStatus", mock.Anything, mock.Anything).Return(nil)
	s.gAuth.On("Validate", mock.Anything, ports.TokenPair{
		AccessToken: ports.TokenPairVal{
			Value: "access_token",
			// Expires: time.Now().Add(time.Hour),
		},
		RefreshToken: ports.TokenPairVal{
			Value: "refresh_token",
			// Expires: time.Now().Add(time.Hour),
		}}).Return(&api.AuthResponse{Result: false, Login: "test123", AccessToken: new(api.Token), RefreshToken: new(api.Token)}, e.ErrAuthFailed)
	s.analyticSender.On("ActionTask", mock.Anything, mock.Anything).Return(nil)

	uuid := "66f5b904-3f54-4da4-ba74-6dfdf8d72efb"
	bodyReq := strings.NewReader("")

	req, err := http.NewRequest("DELETE", fmt.Sprintf("http://localhost:3000/task/v1/tasks/%s", uuid), bodyReq)
	s.NoError(err)
	req.AddCookie(&http.Cookie{Name: "access_token", Value: "access_token"})
	req.AddCookie(&http.Cookie{Name: "refresh_token", Value: "refresh_token"})

	client := http.Client{}
	response, err := client.Do(req)

	// log.Println(err)
	// data, err := ioutil.ReadAll(response.Body)
	// log.Println(string(data))
	s.NoError(err)
	s.Equal(http.StatusForbidden, response.StatusCode)
	response.Body.Close()
}

func (s *authTestSuite) TestDeleteHandlerNotFound() {
	// ctx := context.Background()
	s.emailSender.On("StartEmailWorkers", mock.Anything).Return()
	s.emailSender.On("GetEmailResultChan").Return(make(chan map[models.Email]bool))
	s.db.On("Delete", mock.Anything, "test123", "66f5b904-3f54-4da4-ba74-6dfdf8d72efb").Return(e.ErrNotFound)
	s.db.On("GetMessagesToSend", mock.Anything).Return(map[int]models.KafkaAnalyticMessage{}, nil)
	s.db.On("UpdateMessageStatus", mock.Anything, mock.Anything).Return(nil)
	s.gAuth.On("Validate", mock.Anything, ports.TokenPair{
		AccessToken: ports.TokenPairVal{
			Value: "access_token",
			// Expires: time.Now().Add(time.Hour),
		},
		RefreshToken: ports.TokenPairVal{
			Value: "refresh_token",
			// Expires: time.Now().Add(time.Hour),
		}}).Return(&api.AuthResponse{Result: true, Login: "test123", AccessToken: new(api.Token), RefreshToken: new(api.Token)}, nil)
	s.analyticSender.On("ActionTask", mock.Anything, mock.Anything).Return(nil)

	uuid := "66f5b904-3f54-4da4-ba74-6dfdf8d72efb"
	bodyReq := strings.NewReader("")

	req, err := http.NewRequest("DELETE", fmt.Sprintf("http://localhost:3000/task/v1/tasks/%s", uuid), bodyReq)
	s.NoError(err)
	req.AddCookie(&http.Cookie{Name: "access_token", Value: "access_token"})
	req.AddCookie(&http.Cookie{Name: "refresh_token", Value: "refresh_token"})

	client := http.Client{}
	response, err := client.Do(req)

	// log.Println(err)
	// data, err := ioutil.ReadAll(response.Body)
	// log.Println(string(data))
	s.NoError(err)
	s.Equal(http.StatusNotFound, response.StatusCode)
	response.Body.Close()
}
