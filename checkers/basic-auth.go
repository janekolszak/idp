package checkers

import (
	"../core"

	"fmt"
	"net/http"
)

// Basic Authentication checker.
// Expects Storage to return plain text passwords
type BasicAuth struct {
	PlainTextStorage core.PlainTextStorage
}

func (c BasicAuth) Init() error {
	return nil
}

func (c BasicAuth) Check(r *http.Request) (bool, error) {
	user, pass, ok := r.BasicAuth()
	if !ok {
		return false, fmt.Errorf("Bad Basic Auth format")
	}

	storedPass, err := c.PlainTextStorage.Get(user)
	if err != nil {
		return false, err
	}

	return pass == storedPass, nil
}

func (c BasicAuth) Close() error {
	return nil
}
