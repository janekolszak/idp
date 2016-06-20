package providers

import (
	"github.com/janekolszak/idp/core"
	"net/http"
)

// Well.. at least it's the fastest possible checker implementation..
type Const struct {
	Answer bool
}

func (p Const) Check(*http.Request) error {
	if p.Answer {
		return nil
	}

	return core.ErrorAuthenticationFailure
}

func (s Const) Respond(w http.ResponseWriter, r *http.Request) error {
	http.Error(w, "authorization failed", http.StatusUnauthorized)
	return nil
}
