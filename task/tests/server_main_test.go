//go:build all || integration
// +build all integration

package tests

import (
	"context"
	"fmt"
	"net/http"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"gitlab.com/g6834/team26/task/internal/adapters/grpc"
	h "gitlab.com/g6834/team26/task/internal/adapters/http"
	"gitlab.com/g6834/team26/task/internal/adapters/postgres"
	"gitlab.com/g6834/team26/task/internal/domain/models"
	"gitlab.com/g6834/team26/task/internal/domain/task"
	"gitlab.com/g6834/team26/task/pkg/config"
	"gitlab.com/g6834/team26/task/pkg/logger"
	"gitlab.com/g6834/team26/task/pkg/mocks"
)

type TestcontainersSuite struct {
	suite.Suite

	srv            *h.Server
	pgContainer    testcontainers.Container
	authContainer  testcontainers.Container
	analyticSender *mocks.AnalyticMessageSenderMock
	emailSender    *mocks.EmailSenderMock
	authPort       uint16
}

const (
	dbName = "mtsteta"
	dbUser = "postgres"
	dbPass = "1111"
)

func TestTestcontainers(t *testing.T) {
	suite.Run(t, &TestcontainersSuite{})
}

func (s *TestcontainersSuite) SetupSuite() {
	l := logger.New()
	ctx := context.Background()

	c, err := config.New()
	if err != nil {
		s.Suite.T().Errorf("Error parsing env: %s", err)
	}

	dbInitPath, _ := filepath.Abs("../db.sql")
	dbContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image:        "postgres:14",
			ExposedPorts: []string{"5432"},
			Env: map[string]string{
				"POSTGRES_DB":       dbName,
				"POSTGRES_USER":     dbUser,
				"POSTGRES_PASSWORD": dbPass,
			},
			WaitingFor: wait.ForLog("database system is ready to accept connections"),
			SkipReaper: true,
			AutoRemove: true,
			Mounts: testcontainers.ContainerMounts{
				testcontainers.ContainerMount{
					Source: testcontainers.GenericBindMountSource{
						HostPath: dbInitPath,
					},
					Target:   "/docker-entrypoint-initdb.d/db.sql",
					ReadOnly: false},
			},
		},
		Started: true,
	})
	s.Require().NoError(err)

	time.Sleep(5 * time.Second)

	dbPort, err := dbContainer.MappedPort(ctx, "5432")
	s.Require().NoError(err)
	dbIp, err := dbContainer.Host(ctx)
	s.Require().NoError(err)

	pgConn := fmt.Sprintf("postgres://%s:%s@%s:%d/%s", dbUser, dbPass, dbIp, uint16(dbPort.Int()), dbName)
	db, err := postgres.New(ctx, pgConn)
	if err != nil {
		s.Suite.T().Errorf("db init failed: %s", err)
		s.Suite.T().FailNow()
	}

	contextAuthPath, _ := filepath.Abs("../../auth") // предварительно необходимо склонировать репозиторий auth
	authContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			FromDockerfile: testcontainers.FromDockerfile{
				Context: contextAuthPath,
			},
			ExposedPorts: []string{"2000", "4000"},
			Env: map[string]string{
				"PORT":      "2000",
				"GRPC_PORT": "4000",
			},
			WaitingFor: wait.ForLog("app is started"),
			SkipReaper: true,
			AutoRemove: true,
		},
		Started: true,
	})
	s.Require().NoError(err)

	time.Sleep(5 * time.Second)

	authPort, err := authContainer.MappedPort(ctx, "2000")
	s.Require().NoError(err)
	authGrpcPort, err := authContainer.MappedPort(ctx, "4000")
	s.Require().NoError(err)
	authIp, err := authContainer.Host(ctx)
	s.Require().NoError(err)

	grpcConn := fmt.Sprintf("%s:%d", authIp, uint16(authGrpcPort.Int()))
	grpcAuth, err := grpc.New(grpcConn)
	if err != nil {
		s.Suite.T().Errorf("grpc auth client init failed: %s", err)
		s.Suite.T().FailNow()
	}

	analyticSender := new(mocks.AnalyticMessageSenderMock)
	emailSender := new(mocks.EmailSenderMock)

	taskS := task.New(db, grpcAuth, analyticSender, emailSender)

	srv, err := h.New(l, taskS, c)
	if err != nil {
		s.Suite.T().Errorf("http server creating failed: %s", err)
		s.Suite.T().FailNow()
	}

	s.srv = srv
	s.pgContainer = dbContainer
	s.authContainer = authContainer
	s.analyticSender = analyticSender
	s.emailSender = emailSender
	s.authPort = uint16(authPort.Int())

	go s.srv.Start(ctx)

	s.emailSender.On("StartEmailWorkers", mock.Anything).Return()
	s.emailSender.On("GetEmailResultChan").Return(make(chan map[models.Email]bool))

	s.T().Log("Suite setup is done")
}

func (s *TestcontainersSuite) TearDownSuite() {
	_ = s.srv.Stop(context.Background())
	s.pgContainer.Terminate(context.Background())
	s.authContainer.Terminate(context.Background())
	s.T().Log("Suite stop is done")
}

func (s *TestcontainersSuite) TestDBSelect() {
	// ctx := context.Background()
	// s.gAuthMock.On("Validate", ctx, mock.Anything, mock.Anything).Return(&api.AuthResponse{Result: true, Login: "test123", AccessToken: new(api.Token), RefreshToken: new(api.Token)}, nil)

	// bodyAuthReq := strings.NewReader("{\"login\": \"test123\", \"password\": \"qwerty\"}")
	// reqAuth, err := http.NewRequest("POST", fmt.Sprintf("http://localhost:%d/task/v1/tasks/", s.authPort), bodyAuthReq)
	// s.NoError(err)
	// client := http.Client{}
	// client.Do(reqAuth)
	s.analyticSender.On("ActionTask", mock.Anything, mock.Anything).Return(nil)

	bodyTaskReq := strings.NewReader("")
	reqTask, err := http.NewRequest("GET", "http://localhost:3000/task/v1/tasks/", bodyTaskReq)
	s.NoError(err)

	client := http.Client{}
	responseTask, err := client.Do(reqTask)

	s.NoError(err)
	// s.Equal(http.StatusOK, responseTask.StatusCode)
	s.Equal(http.StatusForbidden, responseTask.StatusCode)
	responseTask.Body.Close()
}
