package checkers

import (
	"net/http"
)

// Well.. at least it's the fastest possible checker implementation..
type Const struct {
	Answer bool
}

func (s Const) Init() error {
	return nil
}

func (s Const) Check(*http.Request) (bool, error) {
	return s.Answer, nil
}

func (s Const) Close() error {
	return nil
}
