package core

import (
	"errors"
)

var (
	ErrorAuthenticationFailure = errors.New("authentication failure")
	ErrorNoSuchUser            = errors.New("no such user")
	ErrorBadRequest            = errors.New("bad request")
	ErrorBadChallengeToken     = errors.New("bad challenge token format")
)
