package core

import (
	"net/http"
)

// Provider validates a http Request and responds to a failed authentication
type Provider interface {
	Check(r *http.Request) (user string, err error)
	Write(w http.ResponseWriter, r *http.Request) error
	WriteError(w http.ResponseWriter, r *http.Request) error
}
