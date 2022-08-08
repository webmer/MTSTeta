package models

import "time"

type Cookie struct {
	Name       string
	Value      string
	Expiration time.Time
}
