package core

import (
	"net/http"
)

// Authenticator validates a http Request and responds to a failed authentication
type Provider interface {
	Check(r *http.Request) error
	Respond(w http.ResponseWriter, r *http.Request) error
}

// type ProviderChain []Provider
