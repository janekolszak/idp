package userdb

type Store interface {
	Check(username, password string) error
	Add(username, password string) error
}
