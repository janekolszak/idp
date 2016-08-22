package core

import (
	"net/http"
)

// Provider validates a http Request and responds to a failed authentication
type Provider interface {

	// Check determines whether the user is authenticated
	Check(r *http.Request) (userid string, err error)

	// Register is called when a new user is being registered
	Register(r *http.Request) (userid string, err error)

	// Writes out the register page
	WriteRegister(w http.ResponseWriter, r *http.Request) error

	// Register is called when a new user is being registered
	Verify(r *http.Request) (userid string, err error)

	// Writes out the verification page
	WriteVerify(w http.ResponseWriter, r *http.Request, userid string) error

	Write(w http.ResponseWriter, r *http.Request) error
	WriteError(w http.ResponseWriter, r *http.Request, err error) error
}
