package core

import (
	"net/http"
)

// Checker validates a http Request.
// How it does this is it's own business.
type Checker interface {
	Init() error
	Check(request *http.Request) (bool, error)
	Close() error
}
