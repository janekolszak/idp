package idp

import (
	"crypto/rsa"
	"net/http"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/gorilla/sessions"
	hclient "github.com/ory-am/hydra/client"
	hjwk "github.com/ory-am/hydra/jwk"
	hoauth2 "github.com/ory-am/hydra/oauth2"
	hydra "github.com/ory-am/hydra/sdk"
	"github.com/patrickmn/go-cache"
)

const (
	VerifyPublicKey   = "VerifyPublic"
	ConsentPrivateKey = "ConsentPrivate"
)

func ClientInfoKey(clientID string) string {
	return "ClientInfo:" + clientID
}

var encryptionkey = "something-very-secret"

// Identity Provider's options
type IDPConfig struct {
	// Client id issued by Hydra
	ClientID string `yaml:"client_id"`

	// Client secret issued by Hydra
	ClientSecret string `yaml:"client_secret"`

	// Hydra's address
	ClusterURL string `yaml:"hydra_address"`

	// Expiration time of internal key cache
	KeyCacheExpiration time.Duration `yaml:"key_cache_expiration"`

	// Expiration time of internal clientid cache
	ClientCacheExpiration time.Duration `yaml:"client_cache_expiration"`

	// Internal cache cleanup interval
	CacheCleanupInterval time.Duration `yaml:"cache_cleanup_interval"`

	// Gorilla sessions Store for storing the Challenge.
	ChallengeStore sessions.Store
}

// Identity Provider helper
type IDP struct {
	config *IDPConfig

	// Communication with Hydra
	hc *hydra.Client

	// Http client for communicating with Hydra
	client *http.Client

	// Cache for all private and public keys
	cache *cache.Cache

	// Prepared cookie options for creating and deleting cookies
	// TODO: Is this the best way to do this?
	createChallengeCookieOptions *sessions.Options
	deleteChallengeCookieOptions *sessions.Options
}

// Create the Identity Provider helper
func NewIDP(config *IDPConfig) *IDP {
	var idp = new(IDP)
	idp.config = config

	// TODO: Pass TTL and refresh period from config
	idp.cache = cache.New(config.KeyCacheExpiration, config.CacheCleanupInterval)
	idp.cache.OnEvicted(func(key string, value interface{}) { idp.refreshCache(key) })

	idp.createChallengeCookieOptions = new(sessions.Options)
	idp.createChallengeCookieOptions.Path = "/"      // TODO: More specific?
	idp.createChallengeCookieOptions.MaxAge = 60 * 5 // 5min
	idp.createChallengeCookieOptions.Secure = false  // TODO: Change to true
	idp.createChallengeCookieOptions.HttpOnly = false

	idp.deleteChallengeCookieOptions = new(sessions.Options)
	idp.deleteChallengeCookieOptions.Path = "/"     // TODO: More specific?
	idp.deleteChallengeCookieOptions.MaxAge = -1    // Mark for deletion
	idp.deleteChallengeCookieOptions.Secure = false // TODO: Change to true
	idp.deleteChallengeCookieOptions.HttpOnly = false

	return idp
}

func (idp *IDP) cacheConsentKey() error {
	consentKey, err := idp.downloadConsentKey()

	duration := cache.DefaultExpiration
	if err != nil {
		// re-cache the result even if there's an error, but
		// do it with a shorter timeout. This will ensure we
		// try to refresh the key once that timeout expires,
		// otherwise we'll _never_ refresh the key again.
		duration = idp.config.CacheCleanupInterval
	}

	idp.cache.Set(ConsentPrivateKey, consentKey, duration)
	return err
}

func (idp *IDP) cacheVerificationKey() error {
	verifyKey, err := idp.downloadVerificationKey()

	duration := cache.DefaultExpiration
	if err != nil {
		// re-cache the result even if there's an error, but
		// do it with a shorter timeout. This will ensure we
		// try to refresh the key once that timeout expires,
		// otherwise we'll _never_ refresh the key again.
		duration = idp.config.CacheCleanupInterval
	}

	idp.cache.Set(VerifyPublicKey, verifyKey, duration)
	return err
}

// Called when any key expires
func (idp *IDP) refreshCache(key string) {
	switch key {
	case VerifyPublicKey:
		idp.cacheVerificationKey()
		return

	case ConsentPrivateKey:
		idp.cacheConsentKey()
		return

	default:
		// Will get here for client IDs.
		// Fine to just let them expire, the next request from that
		// client will trigger a refresh
		return
	}
}

// Downloads the hydra's public key
func (idp *IDP) downloadVerificationKey() (*rsa.PublicKey, error) {

	jwk, err := idp.hc.JWK.GetKey(hoauth2.ConsentChallengeKey, "public")
	if err != nil {
		return nil, err
	}

	rsaKey, ok := hjwk.First(jwk.Keys).Key.(*rsa.PublicKey)
	if !ok {
		return nil, ErrorBadPublicKey
	}

	return rsaKey, nil
}

// Downloads the private key used for signing the consent
func (idp *IDP) downloadConsentKey() (*rsa.PrivateKey, error) {
	jwk, err := idp.hc.JWK.GetKey(hoauth2.ConsentEndpointKey, "private")
	if err != nil {
		return nil, err
	}

	rsaKey, ok := hjwk.First(jwk.Keys).Key.(*rsa.PrivateKey)
	if !ok {
		return nil, ErrorBadPrivateKey
	}

	return rsaKey, nil
}

// Connect to Hydra
func (idp *IDP) Connect(skipTLSVerify bool) error {
	var err error
	if skipTLSVerify {
		idp.hc, err = hydra.Connect(
			hydra.ClientID(idp.config.ClientID),
			hydra.ClientSecret(idp.config.ClientSecret),
			hydra.ClusterURL(idp.config.ClusterURL),
			hydra.SkipTLSVerify(),
		)
	} else {
		idp.hc, err = hydra.Connect(
			hydra.ClientID(idp.config.ClientID),
			hydra.ClientSecret(idp.config.ClientSecret),
			hydra.ClusterURL(idp.config.ClusterURL),
		)
	}

	if err != nil {
		return err
	}

	err = idp.cacheVerificationKey()
	if err != nil {
		return err
	}

	err = idp.cacheConsentKey()
	if err != nil {
		return err
	}

	return nil
}

// Parse and verify the challenge JWT
func (idp *IDP) getChallengeToken(challengeString string) (*jwt.Token, error) {
	token, err := jwt.Parse(challengeString, func(token *jwt.Token) (interface{}, error) {
		_, ok := token.Method.(*jwt.SigningMethodRSA)
		if !ok {
			return nil, ErrorBadSigningMethod
		}

		return idp.getVerificationKey()
	})

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, ErrorInvalidToken
	}

	return token, nil
}

func (idp *IDP) getConsentKey() (*rsa.PrivateKey, error) {
	data, ok := idp.cache.Get(ConsentPrivateKey)
	if !ok {
		return nil, ErrorNotInCache
	}

	key, ok := data.(*rsa.PrivateKey)
	if !ok {
		return nil, ErrorBadKey
	}

	return key, nil
}

func (idp *IDP) getVerificationKey() (*rsa.PublicKey, error) {
	data, ok := idp.cache.Get(VerifyPublicKey)
	if !ok {
		return nil, ErrorNotInCache
	}

	key, ok := data.(*rsa.PublicKey)
	if !ok {
		return nil, ErrorBadKey
	}

	return key, nil
}

func (idp *IDP) getClient(clientID string) (*hclient.Client, error) {
	clientKey := ClientInfoKey(clientID)
	data, ok := idp.cache.Get(clientKey)
	if ok {
		if data != nil {
			client := data.(*hclient.Client)
			return client, nil
		}
		return nil, ErrorNoSuchClient
	}

	client, err := idp.hc.Client.GetClient(clientID)
	if err != nil {
		// Either the client isn't registered in hydra, or maybe hydra is
		// having some problem. Either way, ensure we don't hit hydra again
		// for this client if someone (maybe an attacker) retries quickly.
		idp.cache.Set(clientKey, nil, idp.config.ClientCacheExpiration)
		return nil, err
	}

	c := client.(*hclient.Client)
	idp.cache.Set(clientKey, client, idp.config.ClientCacheExpiration)
	return c, nil
}

// Create a new Challenge. The request will contain all the necessary information from Hydra, passed in the URL.
func (idp *IDP) NewChallenge(r *http.Request, user string) (challenge *Challenge, err error) {
	tokenStr := r.FormValue("challenge")
	if tokenStr == "" {
		// No challenge token
		err = ErrorBadRequest
		return
	}

	token, err := idp.getChallengeToken(tokenStr)
	if err != nil {
		// Most probably, token can't be verified or parsed
		return
	}
	claims := token.Claims.(jwt.MapClaims)

	challenge = new(Challenge)
	challenge.Expires = time.Unix(int64(claims["exp"].(float64)), 0)
	if challenge.Expires.Before(time.Now()) {
		challenge = nil
		err = ErrorChallengeExpired
		return
	}

	// Get data from the challenge jwt
	challenge.Client, err = idp.getClient(claims["aud"].(string))
	if err != nil {
		return nil, err
	}

	challenge.Redirect = claims["redir"].(string)
	challenge.User = user
	challenge.idp = idp

	scopes := claims["scp"].([]interface{})
	challenge.Scopes = make([]string, len(scopes), len(scopes))
	for i, scope := range scopes {
		challenge.Scopes[i] = scope.(string)
	}

	return
}

// Get the Challenge from a cookie, using Gorilla sessions
func (idp *IDP) GetChallenge(r *http.Request) (*Challenge, error) {
	session, err := idp.config.ChallengeStore.Get(r, SessionCookieName)
	if err != nil {
		return nil, err
	}

	challenge, ok := session.Values[SessionCookieName].(*Challenge)
	if !ok {
		return nil, ErrorBadChallengeCookie
	}

	if challenge.Expires.Before(time.Now()) {
		return nil, ErrorChallengeExpired
	}

	challenge.idp = idp

	return challenge, nil
}

// Closes connection to Hydra, cleans cache etc.
func (idp *IDP) Close() {
	idp.client = nil

	// Removes all keys from the cache
	idp.cache.Flush()
}
