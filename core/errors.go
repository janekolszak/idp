package core

import (
	"errors"
)

var (
	ErrorAuthenticationFailure = errors.New("authentication failure")
	ErrorNoSuchUser            = errors.New("no such user")
	ErrorUserAlreadyExists     = errors.New("user already exists")
	ErrorBadRequest            = errors.New("bad request")
	ErrorBadChallengeToken     = errors.New("bad challenge token format")
	ErrorNoChallengeCookie     = errors.New("challenge token isn't stored in a cookie")
	ErrorBadChallengeCookie    = errors.New("bad format of the challenge cookie")
	ErrorChallengeExpired      = errors.New("bad format of the challenge cookie")
	ErrorNoKey                 = errors.New("there's no key in the cache")
	ErrorNoSuchClient          = errors.New("there's no OIDC Client with such id")
	ErrorBadKey                = errors.New("bad key stored in the cache ")
	ErrorInvalidConfig         = errors.New("invalid config")
	ErrorBadPublicKey          = errors.New("cannot conver to public key")
	ErrorBadPrivateKey         = errors.New("cannot conver to private key")
	ErrorNotInCache            = errors.New("cache doesn't have the requested data")
	ErrorSessionExpired        = errors.New("session expired")
)
