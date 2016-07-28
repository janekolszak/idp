package core

import (
	"net/http"
)

// Provider validates a http Request and responds to a failed authentication
type Provider interface {

	// Check determines whether the user is authenticated
	Check(r *http.Request) (user string, err error)

	// Register is called when a new user is being registered
	// Register(r *http.Request) (user string, err error)
	WriteRegister(w http.ResponseWriter, r *http.Request) error

	Write(w http.ResponseWriter, r *http.Request) error
	WriteError(w http.ResponseWriter, r *http.Request, err error) error
}
