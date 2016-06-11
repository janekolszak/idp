package core

import (
	"errors"
)

var (
	ErrorAuthenticationFailure = errors.New("authentication failure")
	ErrorNoSuchUser            = errors.New("no such user")
)
