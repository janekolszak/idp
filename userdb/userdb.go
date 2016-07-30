package userdb

import "time"

type Store interface {
	Check(username, password string) error
	// Add(username, password string) error
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
	Check(username, password string) error
	Insert(user *User, password string) (userid string, err error)
	GetWithID(id string) (user *User, err error)
	GetWithUsername(username string) (user *User, err error)
	SetPasswordWithID(id, password string) error
	Update(user *User) error
	DeleteWithID(id string) error
	SetIsVerifiedWithID(id string) error
}
