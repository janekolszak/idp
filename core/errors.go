package core

import (
	"errors"
)

var (
	ErrorAuthenticationFailure = errors.New("authentication failure")
	ErrorNoSuchUser            = errors.New("no such user")
	ErrorBadRequest            = errors.New("bad request")
	ErrorBadChallengeToken     = errors.New("bad challenge token format")
	ErrorNoChallengeCookie     = errors.New("challenge token isn't stored in a cookie")
	ErrorBadChallengeCookie    = errors.New("bad format of the challenge cookie")
	ErrorChallengeExpired      = errors.New("bad format of the challenge cookie")
	ErrorNoKey                 = errors.New("there's no key in the cache")
	ErrorBadKey                = errors.New("bad key stored in the cache ")
	ErrorBadPublicKey          = errors.New("cannot conver to public key")
	ErrorBadPrivateKey         = errors.New("cannot conver to private key")
)
