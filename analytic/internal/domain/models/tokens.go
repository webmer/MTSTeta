package models

import (
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
