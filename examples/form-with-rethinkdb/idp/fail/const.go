package fail

import (
	"net/http"

	"github.com/janekolszak/idp/core"
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

func (s Const) WriteError(w http.ResponseWriter, r *http.Request, err error) error {
	http.Error(w, "authorization failed", http.StatusUnauthorized)
	return nil
}

func (s Const) Write(w http.ResponseWriter, r *http.Request) error {
	return nil
}
