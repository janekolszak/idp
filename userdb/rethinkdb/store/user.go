package store

import "time"

type User struct {
	ID               string    `json:"id,omitempty" gorethink:"id,omitempty"`
	Username         string    `json:"username" gorethink:"username"`
	Password         []byte    `json:"password,omitempty" gorethink:"password,omitempty"`
	FirstName        string    `json:"firstName" gorethink:"firstName"`
	LastName         string    `json:"lastName" gorethink:"lastName"`
	Email            string    `json:"email" gorethink:"email"`
	IsVerified       bool      `json:"isVerified" gorethink:"isVerified"`
	RegistrationTime time.Time `json:"registrationTime" gorethink:"registrationTime"`
}

func (u *User) GetUsername() string {
	return u.Username
}

func (u *User) GetPassword() []byte {
	return u.Password
}

func (u *User) GetFirstName() string {
	return u.FirstName
}

func (u *User) GetLastName() string {
	return u.LastName
}

func (u *User) GetEmail() string {
	return u.Email
}

func (u *User) GetIsVerified() bool {
	return u.IsVerified
}

func (u *User) GetRegistrationTime() time.Time {
	return u.RegistrationTime
}
