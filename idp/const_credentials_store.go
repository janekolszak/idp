package main

type ConstCredentialsStore struct {
	Answer bool
}

func (s ConstCredentialsStore) Init() error {
	return nil
}

func (s ConstCredentialsStore) Check(user string, password string) (bool, error) {
	return s.Answer, nil
}

func (s ConstCredentialsStore) Close() error {
	return nil
}
