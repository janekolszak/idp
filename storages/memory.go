package main

import (
	"../core"
)

type Memory struct {
	credentials map[string]string
}

func (s Memory) Init() error {
	return nil
}

func (s Memory) Check(r *http.Request) (bool, error) {
	return s.Answer, nil
}

func (s Memory) Close() error {
	return nil
}

func (s Memory) Add(user string, password string) error {
	return
}
