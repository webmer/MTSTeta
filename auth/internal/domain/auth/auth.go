package auth

import (
	"context"
	"errors"
	"github.com/go-chi/jwtauth/v5"
	e "gitlab.com/g6834/team26/auth/internal/domain/errors"
	"gitlab.com/g6834/team26/auth/pkg/config"
	p "gitlab.com/g6834/team26/auth/pkg/password"
	"time"

	"gitlab.com/g6834/team26/auth/internal/domain/models"
	"gitlab.com/g6834/team26/auth/internal/ports"
)

type Service struct {
	db     ports.UserStorage
	config *config.Config
	token  *models.TokenAuth
}

func New(db ports.UserStorage, c *config.Config) *Service {
	return &Service{
		db:     db,
		config: c,
		token: &models.TokenAuth{
			Access:  jwtauth.New("HS256", []byte(c.Server.AccessSecret), nil),
			Refresh: jwtauth.New("HS256", []byte(c.Server.RefreshSecret), nil),
		},
	}
}

/*func (s *Service) Info(ctx context.Context, login string) (*models.User, error) {
	user, err := s.db.Get(ctx, login)
	if err != nil {
		return nil, fmt.Errorf("get user info for login %s failed: %w", login, err)
	}
	return user, nil
}*/

func (s *Service) Validate(ctx context.Context, tokens models.TokenPair) (login string, upd bool, err error) {
	t := time.Now()

	getExp := func(j *jwtauth.JWTAuth, v string) (timeExp time.Time, login string, err error) {
		token, err := jwtauth.VerifyToken(j, v)
		if err != nil {
			return
		}

		exp, ok := token.Get("expires")
		if !ok {
			err = errors.New("not found field expires")
			return
		}

		timeExp, err = time.Parse(time.RFC3339Nano, exp.(string))
		if err != nil {
			err = errors.New("wrong type expires")
			return
		}

		l, ok := token.Get("login")
		if !ok {
			err = errors.New("not found field login")
			return
		}

		login, ok = l.(string)
		if !ok {
			err = errors.New("wrong type login")
			return
		}

		return
	}

	accessExp, lgn, err := getExp(s.token.Access, tokens.AccessToken.Value)
	//if err != nil {
	//	return
	//}
	login = lgn

	if accessExp.Sub(t) <= 0 || err != nil {
		refreshExp, lgn, err := getExp(s.token.Refresh, tokens.RefreshToken.Value)
		if err != nil {
			return "", false, err
		}

		if refreshExp.Sub(t) <= 0 {
			err = e.ErrTokenInvalid
			return "", false, err
		}

		upd = true
		login = lgn
	}

	return login, upd, nil
}

func (s *Service) Login(ctx context.Context, login, password string) (tokens models.TokenPair, err error) {
	u, err := s.db.GetUserByLogin(ctx, login)
	if err != nil {
		return
	}

	if !p.CheckPasswordHash(password, u.Password) {
		err = e.ErrUserInvalid
		return
	}

	tokens, err = s.GenTokens(login)
	if err != nil {
		return
	}

	return
}

func (s *Service) Create(ctx context.Context, login, password string) (err error) {
	pass, err := p.HashPassword(password)
	if err != nil {
		return err
	}

	err = s.db.Create(ctx, login, pass)
	if err != nil {
		return
	}

	return
}

func (s *Service) GenTokens(login string) (tokens models.TokenPair, err error) {
	tAccess := time.Now().Add(time.Minute)
	tRefresh := time.Now().Add(time.Hour)

	tokenClaim := map[string]interface{}{}
	tokenClaim["login"] = login
	tokenClaim["expires"] = tAccess

	_, tokenAc, err := s.token.Access.Encode(tokenClaim)
	if err != nil {
		return
	}

	tokenClaim["expires"] = tRefresh
	_, tokenRe, err := s.token.Refresh.Encode(tokenClaim)
	if err != nil {
		return
	}

	tokens = models.TokenPair{
		AccessToken: models.TokenPairVal{
			Value:   tokenAc,
			Expires: tAccess,
		},
		RefreshToken: models.TokenPairVal{
			Value:   tokenRe,
			Expires: tRefresh,
		},
	}

	return
}
