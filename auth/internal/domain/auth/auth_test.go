//go:build auth_domain || all
// +build auth_domain all

package auth

import (
	"gitlab.com/g6834/team26/auth/internal/adapters/mongo"
	"gitlab.com/g6834/team26/auth/internal/domain/models"
	"gitlab.com/g6834/team26/auth/pkg/config"
	"golang.org/x/net/context"
	"testing"
	"time"
)

func TestAuth(t *testing.T) {
	c, _ := config.New()
	s := New(&mongo.Database{}, c)

	t.Run("test func Validate - ok", func(t *testing.T) {
		tm := time.Now()

		_, normalTokenAc, err := s.token.Access.Encode(map[string]interface{}{"login": "testNormalLogin", "expires": tm.Add(time.Minute)})
		if err != nil {
			t.Errorf("error gen access token %d", err)
		}
		_, normalTokenRe, err := s.token.Refresh.Encode(map[string]interface{}{"login": "testNormalLogin", "expires": tm.Add(time.Hour)})
		if err != nil {
			t.Errorf("error gen refresh token %d", err)
		}

		tokens := models.TokenPair{
			AccessToken:  models.TokenPairVal{Value: normalTokenAc},
			RefreshToken: models.TokenPairVal{Value: normalTokenRe},
		}

		l, _, err := s.Validate(context.Background(), tokens)
		if l != "testNormalLogin" {
			t.Errorf("not valide %s", l)
		}
	})
	t.Run("test func Validate - refresh tokens", func(t *testing.T) {
		tm := time.Now()

		_, normalTokenAc, err := s.token.Access.Encode(map[string]interface{}{"login": "testNormalLogin", "expires": tm.Add(-time.Minute)})
		if err != nil {
			t.Errorf("error gen access token %d", err)
		}
		_, normalTokenRe, err := s.token.Refresh.Encode(map[string]interface{}{"login": "testNormalLogin", "expires": tm.Add(time.Hour)})
		if err != nil {
			t.Errorf("error gen refresh token %d", err)
		}

		tokens := models.TokenPair{
			AccessToken:  models.TokenPairVal{Value: normalTokenAc},
			RefreshToken: models.TokenPairVal{Value: normalTokenRe},
		}

		_, u, err := s.Validate(context.Background(), tokens)
		if !u {
			t.Errorf("error parsing time to refresh")
		}
	})
	t.Run("test func Validate - empty access need refresh tokens", func(t *testing.T) {
		tm := time.Now()

		normalTokenAc := ""
		_, normalTokenRe, err := s.token.Refresh.Encode(map[string]interface{}{"login": "testNormalLogin", "expires": tm.Add(time.Hour)})
		if err != nil {
			t.Errorf("error gen refresh token %d", err)
		}

		tokens := models.TokenPair{
			AccessToken:  models.TokenPairVal{Value: normalTokenAc},
			RefreshToken: models.TokenPairVal{Value: normalTokenRe},
		}

		_, u, err := s.Validate(context.Background(), tokens)
		if !u {
			t.Errorf("error parsing time to refresh")
		}
	})
	t.Run("test func Validate - empty tokens", func(t *testing.T) {
		normalTokenAc := ""
		normalTokenRe := ""

		tokens := models.TokenPair{
			AccessToken:  models.TokenPairVal{Value: normalTokenAc},
			RefreshToken: models.TokenPairVal{Value: normalTokenRe},
		}

		_, _, err := s.Validate(context.Background(), tokens)
		if err == nil {
			t.Errorf("empty tokens valid")
		}
	})
	t.Run("test func Validate - expires has passed", func(t *testing.T) {
		tm := time.Now()

		_, normalTokenAc, err := s.token.Access.Encode(map[string]interface{}{"login": "testNormalLogin", "expires": tm.Add(-time.Minute)})
		if err != nil {
			t.Errorf("error gen access token %d", err)
		}
		_, normalTokenRe, err := s.token.Refresh.Encode(map[string]interface{}{"login": "testNormalLogin", "expires": tm.Add(-time.Hour)})
		if err != nil {
			t.Errorf("error gen refresh token %d", err)
		}

		tokens := models.TokenPair{
			AccessToken:  models.TokenPairVal{Value: normalTokenAc},
			RefreshToken: models.TokenPairVal{Value: normalTokenRe},
		}

		_, _, err = s.Validate(context.Background(), tokens)
		if err == nil {
			t.Errorf("tokens valid")
		}
	})
}
