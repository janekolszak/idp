package idp

import (
	"errors"
)

var (
	ErrorBadPublicKey       = errors.New("cannot convert to public key")
	ErrorBadPrivateKey      = errors.New("cannot convert to private key")
	ErrorBadRequest         = errors.New("bad request")
	ErrorBadChallengeCookie = errors.New("bad format of the challenge cookie")
	ErrorChallengeExpired   = errors.New("challenge expired")
	ErrorNoSuchClient       = errors.New("there's no OIDC Client with such id")
	ErrorBadKey             = errors.New("bad key stored in the cache ")
	ErrorNotInCache         = errors.New("cache doesn't have the requested data")
	ErrorBadSigningMethod   = errors.New("bad signing method")
	ErrorInvalidToken       = errors.New("invalid token")
)
