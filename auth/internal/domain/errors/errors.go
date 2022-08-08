package errors

import "errors"

var (
	ErrNotFound     = errors.New("not found")
	ErrTokenInvalid = errors.New("invalid token")
	ErrUserInvalid  = errors.New("not valid user")
	ErrUserNotFound = errors.New("user not found")
)
