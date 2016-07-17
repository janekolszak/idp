package userdb

import "time"

type Store interface {
	Check(username, password string) error
	Add(username, password string) error
}

type UserInfo interface {
	GetUsername() string
	GetPassword() string
	GetFirstName() string
	GetLastName() string
	GetEmail() string
	GetIsVerified() bool
	GetRegistrationTime() time.Time
}

type UserStore interface {
	Get(username string) (UserInfo, error)
	Check(username, password string) error
	Insert(userinfo UserInfo) error
	Update(userinfo UserInfo) error
	Delete(username string) error
}
