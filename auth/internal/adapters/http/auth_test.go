//go:build e2e || all
// +build e2e all

package http

import (
	"encoding/json"
	"fmt"
	"github.com/go-chi/jwtauth/v5"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/suite"
	"gitlab.com/g6834/team26/auth/internal/adapters/mongo"
	"gitlab.com/g6834/team26/auth/internal/domain/auth"
	"gitlab.com/g6834/team26/auth/internal/domain/models"
	"gitlab.com/g6834/team26/auth/pkg/config"
	"golang.org/x/net/context"
	gHttp "net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"
)

type authTestSuite struct {
	suite.Suite

	srv *Server
	db  *mongo.Database
	l   zerolog.Logger
	c   *config.Config
	t   *models.TokenAuth
}

func TestAuthTestSuite(t *testing.T) {
	suite.Run(t, &authTestSuite{})
}

func (s *authTestSuite) SetupSuite() {
	s.l = zerolog.Nop()

	c, err := config.New()
	if err != nil {
		s.Suite.T().Errorf("error parsing env: %v", err)
	}
	c.Server.Port = "3010"
	s.c = c

	s.db, err = mongo.New(context.Background(), c.Server.AuthorizationDataBaseConnectionString)
	if err != nil {
		s.Suite.T().Errorf("db init failed: %s", err)
	}
	authS := auth.New(s.db, c)

	s.srv, err = New(&s.l, authS, c)
	if err != nil {
		s.Suite.T().Errorf("http server creating failed: %s", err)
	}

	s.t = &models.TokenAuth{
		Access:  jwtauth.New("HS256", []byte(c.Server.AccessSecret), nil),
		Refresh: jwtauth.New("HS256", []byte(c.Server.RefreshSecret), nil),
	}

	go s.srv.Start()
}

func (s *authTestSuite) TearDownSuite() {
	_ = s.srv.Stop(context.Background())
	_ = s.db.Disconnect(context.Background())
}

func (s *authTestSuite) TestUserLogin() {
	login := "test123"
	password := "qwerty"

	bodyReq := strings.NewReader(fmt.Sprintf("{\"login\":\"%s\", \"password\":\"%s\"}", login, password))
	req, err := gHttp.NewRequest("POST", fmt.Sprintf("http://localhost:%s/auth/v1/login", s.c.Server.Port), bodyReq)
	s.NoError(err)

	client := gHttp.Client{}
	response, err := client.Do(req)

	s.NoError(err)
	s.Equal(gHttp.StatusOK, response.StatusCode)

	var authResp models.AuthResponse

	s.NoError(json.NewDecoder(response.Body).Decode(&authResp))

	s.Equal(authResp.Status, "ok")

	response.Body.Close()
}

func (s *authTestSuite) TestValidate() {
	tm := time.Now()

	_, normalTokenAc, err := s.t.Access.Encode(map[string]interface{}{"login": "testNormalLogin", "expires": tm.Add(time.Minute)})
	if err != nil {
		s.Suite.T().Errorf("error gen access token %d", err)
	}
	_, normalTokenRe, err := s.t.Refresh.Encode(map[string]interface{}{"login": "testNormalLogin", "expires": tm.Add(time.Hour)})
	if err != nil {
		s.Suite.T().Errorf("error gen refresh token %d", err)
	}

	recorder := httptest.NewRecorder()

	gHttp.SetCookie(recorder, &gHttp.Cookie{Name: s.c.Server.AccessCookie, Value: normalTokenAc})
	gHttp.SetCookie(recorder, &gHttp.Cookie{Name: s.c.Server.RefreshCookie, Value: normalTokenRe})

	u, err := url.Parse(fmt.Sprintf("http://localhost:%s/auth/v1/i", s.c.Server.Port))
	if err != nil {
		s.Suite.T().Errorf("error parse url %d", err)
	}

	request := &gHttp.Request{Method: gHttp.MethodPost, URL: u, Header: gHttp.Header{"Cookie": recorder.HeaderMap["Set-Cookie"]}}

	//request.AddCookie(&gHttp.Cookie{Name: s.c.Server.AccessCookie, Value: normalTokenAc})
	//request.AddCookie(&gHttp.Cookie{Name: s.c.Server.RefreshCookie, Value: normalTokenRe})

	client := gHttp.Client{}
	response, err := client.Do(request)

	s.NoError(err)
	s.Equal(gHttp.StatusOK, response.StatusCode)

	var authResp models.AuthResponse

	s.NoError(json.NewDecoder(response.Body).Decode(&authResp))

	s.Equal(authResp.Status, "ok")
	s.Equal(authResp.Login, "testNormalLogin")

	response.Body.Close()
}
