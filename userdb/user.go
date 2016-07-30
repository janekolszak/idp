package userdb

import "time"

type User struct {
	ID               string    `schema:"-"         json:"id,omitempty"       gorethink:"id,omitempty"`
	Username         string    `schema:"username"  json:"username"           gorethink:"username"`
	Password         []byte    `schema:"password"  json:"password,omitempty" gorethink:"password,omitempty"`
	FirstName        string    `schema:"firstName" json:"firstName"          gorethink:"firstName"`
	LastName         string    `schema:"lastName"  json:"lastName"           gorethink:"lastName"`
	Email            string    `schema:"email"     json:"email"              gorethink:"email"`
	IsVerified       bool      `schema:"-"         json:"isVerified"         gorethink:"isVerified"`
	RegistrationTime time.Time `schema:"-"         json:"registrationTime"   gorethink:"registrationTime"`
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
