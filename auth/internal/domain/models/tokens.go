package models

import (
	"github.com/go-chi/jwtauth/v5"
	"time"
)

type TokenPair struct {
	AccessToken  TokenPairVal
	RefreshToken TokenPairVal
}

type TokenPairVal struct {
	Value   string
	Expires time.Time
}

type TokenAuth struct {
	Access  *jwtauth.JWTAuth
	Refresh *jwtauth.JWTAuth
}
